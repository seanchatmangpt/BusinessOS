package ontology

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	bossemconv "github.com/rhl/businessos-backend/internal/semconv"
)

const (
	// l0NamedGraph is the canonical Oxigraph named graph for L0 ground-truth facts.
	// L1-L3 SPARQL files reference this graph via FROM <l0NamedGraph>.
	l0NamedGraph   = "https://chatmangpt.com/ontology/businessos/l0"
	l0SyncInterval = 15 * time.Minute
	l0SubprocTimeout = 30 * time.Second
)

// bosExecuteResult mirrors the JSON stdout from `bos ontology execute`.
type bosExecuteResult struct {
	TotalRows            int `json:"total_rows"`
	TotalConstructTriples int `json:"total_construct_triples"`
	Tables               []struct {
		Table             string `json:"table"`
		RowsLoaded        int    `json:"rows_loaded"`
		TriplesGenerated  int    `json:"triples_generated"`
		ConstructTriples  int    `json:"construct_triples"`
	} `json:"tables"`
}

// CanopyIntelligencePusher is an optional side-car that receives the L0 sync
// results and forwards them to Canopy.  Implemented in integrations/canopy.
type CanopyIntelligencePusher interface {
	PushIntelligence(ctx context.Context, caseCount, handoffCount int) error
}

// BoardchairL0Sync syncs BusinessOS case data into Oxigraph as L0 RDF facts
// by invoking the bos CLI as a subprocess.
//
// WvdA: L0 = ground truth event log. Must be continuously updated.
// Armstrong: supervised background job, subprocess bounded by l0SubprocTimeout.
type BoardchairL0Sync struct {
	bosPath     string // path to bos binary (env BOS_PATH)
	mappingFile string // path to mapping JSON (env BOS_MAPPING_FILE)
	dbURL       string // PostgreSQL connection string (env DATABASE_URL)
	tracer      trace.Tracer
	ticker      *time.Ticker
	done        chan struct{}
	canopy      CanopyIntelligencePusher // nil → skip push
}

// NewBoardchairL0Sync creates a new L0 sync service.
//
// bosPath: path to the `bos` binary (default "bos").
// mappingFile: path to the mapping JSON file for `bos ontology execute`.
// dbURL: PostgreSQL connection string.
func NewBoardchairL0Sync(bosPath, mappingFile, dbURL string) *BoardchairL0Sync {
	if bosPath == "" {
		bosPath = "bos"
	}
	return &BoardchairL0Sync{
		bosPath:     bosPath,
		mappingFile: mappingFile,
		dbURL:       dbURL,
		tracer:      otel.Tracer("businessos.board"),
		done:        make(chan struct{}),
	}
}

// SetCanopyPusher wires an optional Canopy intelligence push after each sync.
func (s *BoardchairL0Sync) SetCanopyPusher(p CanopyIntelligencePusher) {
	s.canopy = p
}

// Start begins the periodic L0 sync. Call in a goroutine.
// Armstrong: caller must supervise — if this crashes, restart it.
func (s *BoardchairL0Sync) Start(ctx context.Context) {
	s.ticker = time.NewTicker(l0SyncInterval)
	defer s.ticker.Stop()

	// Run immediately on start
	if err := s.Sync(ctx); err != nil {
		slog.Error("board.l0_sync initial sync failed", "error", err)
	}

	for {
		select {
		case <-s.ticker.C:
			if err := s.Sync(ctx); err != nil {
				slog.Error("board.l0_sync periodic sync failed", "error", err)
			}
		case <-s.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop halts the sync loop.
func (s *BoardchairL0Sync) Stop() {
	close(s.done)
}

// Sync performs one L0 sync by invoking `bos ontology execute` as a subprocess.
// The subprocess writes N-Triples directly to Oxigraph (port 7878); Go never
// touches Oxigraph directly.
func (s *BoardchairL0Sync) Sync(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, bossemconv.BoardL0SyncSpan)
	defer span.End()

	if s.mappingFile == "" {
		slog.Warn("board.l0_sync skipped: BOS_MAPPING_FILE not set")
		return nil
	}

	// Inject W3C traceparent so bos CLI participates in the same trace.
	traceparent := extractTraceparent(ctx)

	// WvdA: 30s hard timeout on the subprocess — deadlock freedom guarantee.
	subCtx, cancel := context.WithTimeout(ctx, l0SubprocTimeout)
	defer cancel()

	cmd := exec.CommandContext(subCtx, s.bosPath,
		"ontology", "execute",
		"--mapping", s.mappingFile,
		"--database", s.dbURL,
		"--graph", l0NamedGraph,
	)
	cmd.Env = append(os.Environ(),
		"TRACEPARENT="+traceparent,
		"WEAVER_LIVE_CHECK="+os.Getenv("WEAVER_LIVE_CHECK"),
		"WEAVER_OTLP_ENDPOINT="+os.Getenv("WEAVER_OTLP_ENDPOINT"),
		"CHATMANGPT_CORRELATION_ID="+os.Getenv("CHATMANGPT_CORRELATION_ID"),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", err)))
		slog.Error("board.l0_sync bos subprocess failed",
			"error", err, "stderr", stderr.String())
		return fmt.Errorf("board.l0_sync bos execute: %w", err)
	}

	// Parse JSON output from bos CLI to get counts for span attributes.
	var result bosExecuteResult
	caseCount, handoffCount := 0, 0
	if err := json.Unmarshal(stdout.Bytes(), &result); err == nil {
		caseCount = result.TotalRows
		handoffCount = result.TotalConstructTriples
		// Attempt to split by table name if available.
		for _, t := range result.Tables {
			switch t.Table {
			case "cases":
				caseCount = t.RowsLoaded
			case "process_handoffs":
				handoffCount = t.RowsLoaded
			}
		}
	}

	span.SetAttributes(
		attribute.Int("board.l0_sync.case_count", caseCount),
		attribute.Int("board.l0_sync.handoff_count", handoffCount),
	)
	slog.Info("board.l0_sync complete", "cases", caseCount, "handoffs", handoffCount)

	// Fire-and-forget push to Canopy (Armstrong: never let failures propagate).
	if s.canopy != nil {
		go func() {
			pushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.canopy.PushIntelligence(pushCtx, caseCount, handoffCount); err != nil {
				slog.Warn("board.l0_sync canopy push failed", "error", err)
			}
		}()
	}

	return nil
}

// extractTraceparent reads the W3C traceparent from the current span context.
func extractTraceparent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if tp, ok := carrier["traceparent"]; ok {
		return tp
	}
	return ""
}
