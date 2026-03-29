//! bos — BusinessOS data layer CLI.
//!
//! Noun-verb command structure for ODCS workspace operations,
//! schema management, data pipelines, decision records, and knowledge base.
//!
//! ## Examples
//!
//! ```bash
//! bos workspace init --name my-project
//! bos workspace validate --path ./my-project
//! bos schema convert --input schema.sql --output schema.odc
//! bos decisions list
//! bos knowledge index --directory ./docs
//! bos ontology construct --mapping mappings.json
//! bos ontology execute --mapping mappings.json --database $DATABASE_URL
//! ```

mod nouns;
mod telemetry;

fn main() -> clap_noun_verb::Result<()> {
    // When WEAVER_LIVE_CHECK=true, telemetry::init_otel() installs the full
    // subscriber (fmt + OTLP gRPC).  Otherwise install the plain fmt layer
    // here and return a no-op guard.
    let _otel = if std::env::var("WEAVER_LIVE_CHECK")
        .map(|v| v.eq_ignore_ascii_case("true"))
        .unwrap_or(false)
    {
        telemetry::init_otel()
    } else {
        tracing_subscriber::fmt()
            .with_env_filter(
                tracing_subscriber::EnvFilter::from_default_env()
                    .add_directive("bos=info".parse().unwrap()),
            )
            .without_time()
            .init();
        telemetry::OtelGuard::noop()
    };

    clap_noun_verb::run()
}
