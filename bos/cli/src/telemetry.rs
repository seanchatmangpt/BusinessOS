//! OTEL initialisation for the bos CLI.
//!
//! When `WEAVER_LIVE_CHECK=true` is set, the CLI exports spans to an OTLP
//! gRPC endpoint (default `http://localhost:4317`, override via
//! `WEAVER_OTLP_ENDPOINT`).  In all other cases the global tracer is a no-op
//! so there is zero overhead in normal development use.
//!
//! The returned `OtelGuard` must be held for the lifetime of `main()`.  When
//! it is dropped the exporter flushes all pending spans before the process
//! exits.

use opentelemetry::global;
use opentelemetry::trace::TracerProvider as _;
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::{runtime, trace as sdktrace, Resource};
use opentelemetry::KeyValue;
use tracing_opentelemetry::OpenTelemetryLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

/// Holds the active tracer provider so that `Drop` can flush spans.
pub struct OtelGuard {
    /// Present only when WEAVER_LIVE_CHECK is enabled.
    provider: Option<sdktrace::SdkTracerProvider>,
}

impl OtelGuard {
    /// Return a no-op guard for when OTEL is disabled.
    /// Does not install any subscriber.
    pub fn noop() -> Self {
        Self { provider: None }
    }
}

impl Drop for OtelGuard {
    fn drop(&mut self) {
        if let Some(ref provider) = self.provider {
            if let Err(e) = provider.shutdown() {
                eprintln!("[bos telemetry] flush error: {e}");
            }
        }
        global::shutdown_tracer_provider();
    }
}

/// Initialise OTEL (or no-op) and wire the tracing-opentelemetry bridge.
///
/// Call once at the top of `main()` and hold the guard for the entire
/// process lifetime.
pub fn init_otel() -> OtelGuard {
    let live_check = std::env::var("WEAVER_LIVE_CHECK")
        .map(|v| v.eq_ignore_ascii_case("true"))
        .unwrap_or(false);

    if live_check {
        let endpoint = std::env::var("WEAVER_OTLP_ENDPOINT")
            .unwrap_or_else(|_| "http://localhost:4317".to_string());

        let correlation_id = std::env::var("CHATMANGPT_CORRELATION_ID")
            .unwrap_or_else(|_| uuid::Uuid::new_v4().to_string());

        let resource = Resource::new(vec![
            KeyValue::new("service.name", "bos-cli"),
            KeyValue::new("chatmangpt.run.correlation_id", correlation_id),
        ]);

        let exporter = opentelemetry_otlp::SpanExporter::builder()
            .with_tonic()
            .with_endpoint(endpoint)
            .build()
            .expect("failed to build OTLP exporter");

        let provider = sdktrace::SdkTracerProvider::builder()
            .with_batch_exporter(exporter, runtime::Tokio)
            .with_resource(resource)
            .build();

        let tracer = provider.tracer("bos-cli");
        global::set_tracer_provider(provider.clone());

        let env_filter = EnvFilter::from_default_env()
            .add_directive("bos=info".parse().unwrap());

        tracing_subscriber::registry()
            .with(env_filter)
            .with(tracing_subscriber::fmt::layer().without_time())
            .with(OpenTelemetryLayer::new(tracer))
            .init();

        OtelGuard { provider: Some(provider) }
    } else {
        // No-op: use the plain fmt subscriber that main() already installs
        OtelGuard { provider: None }
    }
}
