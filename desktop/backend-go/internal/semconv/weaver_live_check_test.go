package semconv

import (
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	bosotel "github.com/rhl/businessos-backend/internal/otel"
)

func TestMain(m *testing.M) {
	bosotel.TestMainFunc(m)
}

func TestWeaverLiveCheckEmitsCorrelationSpan(t *testing.T) {
	if os.Getenv(bosotel.EnvWeaverLiveCheck) != "true" {
		t.Skip("set WEAVER_LIVE_CHECK=true for Weaver live-check")
	}
	cid := os.Getenv("CHATMANGPT_CORRELATION_ID")
	if cid == "" {
		t.Fatal("CHATMANGPT_CORRELATION_ID must be set for live-check correlation")
	}
	ctx := context.Background()
	tr := otel.Tracer("businessos")
	_, sp := tr.Start(ctx, SpanNameBosComplianceCheck)
	sp.SetAttributes(
		attribute.String("chatmangpt.run.correlation_id", cid),
		attribute.String("compliance.framework", "soc2"),
	)
	sp.End()
}
