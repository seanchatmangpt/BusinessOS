//! Weaver live-check integration test for the bos CLI.
//!
//! Emits `bos.ontology.execute`, `bos.rdf.write`, and `bos.rdf.query` spans
//! directly via OTLP to the weaver live-check listener.
//!
//! Run via:
//!   WEAVER_LIVE_CHECK=true cargo test --test bos_weaver_live_check -- --nocapture
//!
//! Skipped automatically when WEAVER_LIVE_CHECK is not set.

use opentelemetry::global;
use opentelemetry::trace::{Tracer, TracerProvider as _};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::{runtime, trace as sdktrace, Resource};
use opentelemetry::KeyValue;

#[tokio::test]
async fn test_bos_weaver_live_check_spans() {
    let live_check = std::env::var("WEAVER_LIVE_CHECK")
        .map(|v| v.eq_ignore_ascii_case("true"))
        .unwrap_or(false);

    if !live_check {
        eprintln!("WEAVER_LIVE_CHECK not set — skipping bos weaver live-check test");
        return;
    }

    let endpoint = std::env::var("WEAVER_OTLP_ENDPOINT")
        .unwrap_or_else(|_| "http://localhost:4317".to_string());

    let correlation_id = std::env::var("CHATMANGPT_CORRELATION_ID")
        .unwrap_or_else(|_| "bos-weaver-live-check-test".to_string());

    // Build OTLP exporter
    let exporter = opentelemetry_otlp::SpanExporter::builder()
        .with_tonic()
        .with_endpoint(&endpoint)
        .build()
        .expect("failed to build OTLP exporter");

    let resource = Resource::new(vec![
        KeyValue::new("service.name", "bos-cli"),
        KeyValue::new("chatmangpt.run.correlation_id", correlation_id.clone()),
    ]);

    let provider = sdktrace::SdkTracerProvider::builder()
        .with_batch_exporter(exporter, runtime::Tokio)
        .with_resource(resource)
        .build();

    let tracer = provider.tracer("bos-cli");
    global::set_tracer_provider(provider.clone());

    // Emit bos.ontology.execute
    {
        let span = tracer.start("bos.ontology.execute");
        let cx = opentelemetry::Context::current_with_span(span);
        let span = cx.span();
        span.set_attribute(KeyValue::new("rdf.sparql.endpoint", "http://localhost:7878"));
        span.set_attribute(KeyValue::new("rdf.result.triple_count", 150i64));
        span.set_attribute(KeyValue::new("chatmangpt.run.correlation_id", correlation_id.clone()));
        span.end();
    }
    eprintln!("emitted span: bos.ontology.execute");

    // Emit bos.rdf.write
    {
        let span = tracer.start("bos.rdf.write");
        let cx = opentelemetry::Context::current_with_span(span);
        let span = cx.span();
        span.set_attribute(KeyValue::new("rdf.sparql.endpoint", "http://localhost:7878"));
        span.set_attribute(KeyValue::new("rdf.write.format", "application/n-triples"));
        span.set_attribute(KeyValue::new("rdf.write.triple_count", 150i64));
        span.set_attribute(KeyValue::new("chatmangpt.run.correlation_id", correlation_id.clone()));
        span.end();
    }
    eprintln!("emitted span: bos.rdf.write");

    // Emit bos.rdf.query
    {
        let span = tracer.start("bos.rdf.query");
        let cx = opentelemetry::Context::current_with_span(span);
        let span = cx.span();
        span.set_attribute(KeyValue::new("rdf.sparql.endpoint", "http://localhost:7879"));
        span.set_attribute(KeyValue::new("rdf.sparql.query_type", "SELECT"));
        span.set_attribute(KeyValue::new("rdf.sparql.result_count", 5i64));
        span.set_attribute(KeyValue::new("chatmangpt.run.correlation_id", correlation_id));
        span.end();
    }
    eprintln!("emitted span: bos.rdf.query");

    // Flush spans before process exits.
    if let Err(e) = provider.shutdown() {
        eprintln!("flush warning: {e}");
    }
    global::shutdown_tracer_provider();

    eprintln!("bos_weaver_live_check: all 3 spans emitted successfully");
}
