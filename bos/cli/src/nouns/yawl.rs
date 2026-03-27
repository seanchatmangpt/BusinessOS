use clap_noun_verb_macros::{noun, verb};
use clap_noun_verb::Result;
use serde::Serialize;

#[derive(Serialize)]
#[serde(rename_all = "snake_case")]
pub struct YawlServing {
    pub host: String,
    pub port: u16,
    pub url: String,
}

#[noun("yawl", "YAWL engine connector — serve HTTP API for health, cases, and process discovery")]

/// Serve the YAWL HTTP API
///
/// Starts a local HTTP server exposing:
///   GET  /api/yawl/health   — engine reachability ({"status":"up"|"down"})
///   GET  /api/yawl/cases    — list running cases as JSON array
///   POST /api/yawl/discover — body: {"spec_id":"..."} → ProcessDiscoveryResult
///
/// The server reads YAWL_ENGINE_URL from the environment (default: http://localhost:8080).
///
/// # Arguments
/// * `host` - Host to bind [default: 127.0.0.1]
/// * `port` - Port to listen on [default: 8090]
#[verb("serve")]
fn serve(host: Option<String>, port: Option<u16>) -> Result<YawlServing> {
    let h = host.unwrap_or_else(|| "127.0.0.1".to_string());
    let p = port.unwrap_or(8090);
    let url = format!("http://{}:{}/api/yawl", h, p);

    let config = bos_core::YawlServeConfig {
        host: h.clone(),
        port: p,
    };

    // serve_yawl_api blocks until the process is killed — this is intentional
    // (same behaviour as `bos ontology serve`).
    bos_core::serve_yawl_api(config)
        .map_err(|e| clap_noun_verb::NounVerbError::execution_error(e.to_string()))?;

    // Unreachable in practice; satisfies the return type.
    Ok(YawlServing {
        host: h,
        port: p,
        url,
    })
}
