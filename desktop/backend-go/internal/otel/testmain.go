package otel

import (
	"context"
	"fmt"
	"os"
	"time"
)

// TestMainFunc is a drop-in TestMain implementation that conditionally sets up
// Weaver live-check. Package test files should call this from their TestMain:
//
//	func TestMain(m *testing.M) {
//	    otel.TestMainFunc(m)
//	}
//
// When WEAVER_LIVE_CHECK=true, it configures the global TracerProvider to
// export spans to the Weaver OTLP receiver, ensuring live-check validation
// during test execution.
func TestMainFunc(m interface{ Run() int }) {
	if IsLiveCheckEnabled() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdown, err := SetupWeaverLiveCheck(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "weaver live-check setup failed: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "weaver shutdown warning: %v\n", err)
			}
		}()
	}

	os.Exit(m.Run())
}
