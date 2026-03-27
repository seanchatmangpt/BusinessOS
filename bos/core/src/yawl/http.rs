//! HTTP API server for the YAWL connector.
//!
//! Exposes three routes under `/api/yawl/`:
//!
//! | Method | Path                | Description                              |
//! |--------|---------------------|------------------------------------------|
//! | GET    | /api/yawl/health    | Engine reachability check                |
//! | GET    | /api/yawl/cases     | List all running YAWL cases              |
//! | POST   | /api/yawl/discover  | Discover process model from YAWL log     |
//!
//! The server follows the same raw-TCP pattern as `ontology::serve` so no
//! additional dependencies are required.

use super::YawlConnector;
use anyhow::{Context, Result};
use std::net::TcpListener;
use tracing::{debug, info, warn};

/// Configuration for the YAWL HTTP API server.
#[derive(Debug, Clone)]
pub struct YawlServeConfig {
    /// Host to bind to [default: 127.0.0.1]
    pub host: String,
    /// Port to listen on [default: 8090]
    pub port: u16,
}

impl Default for YawlServeConfig {
    fn default() -> Self {
        Self {
            host: "127.0.0.1".to_string(),
            port: 8090,
        }
    }
}

/// Start the YAWL HTTP API server and block until the process is killed.
///
/// The `connector` is cloned for every incoming connection so it is shared
/// safely across the single-threaded accept loop (the same model used by the
/// ontology serve module).
pub fn serve(config: YawlServeConfig) -> Result<()> {
    let addr = format!("{}:{}", config.host, config.port);
    let listener =
        TcpListener::bind(&addr).with_context(|| format!("Failed to bind to {}", addr))?;

    info!(
        "YAWL API server listening on http://{}/api/yawl/{{health,cases,discover}}",
        addr
    );
    eprintln!(
        "YAWL API server listening on http://{}/api/yawl/{{health,cases,discover}}",
        addr
    );
    eprintln!("Press Ctrl+C to stop.");

    // Build a single-threaded async runtime so we can await YawlConnector methods
    // inside the synchronous accept loop — same approach as the rest of bos-core.
    let rt = tokio::runtime::Builder::new_current_thread()
        .enable_all()
        .build()
        .context("Failed to build Tokio runtime")?;

    let connector = YawlConnector::from_env();

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                let conn_result = rt.block_on(handle_connection(&connector, stream));
                if let Err(e) = conn_result {
                    debug!("Connection error: {}", e);
                }
            }
            Err(e) => {
                debug!("Accept error: {}", e);
            }
        }
    }

    Ok(())
}

// ---------------------------------------------------------------------------
// Internal connection handler
// ---------------------------------------------------------------------------

async fn handle_connection(
    connector: &YawlConnector,
    stream: std::net::TcpStream,
) -> Result<()> {
    use std::io::{BufRead, Write};

    let _ = stream.set_read_timeout(Some(std::time::Duration::from_secs(10)));
    let _ = stream.set_write_timeout(Some(std::time::Duration::from_secs(10)));

    // We need simultaneous read + write on the same TcpStream.  Clone the stream
    // for writing so BufReader can own the read half.
    let write_stream = stream.try_clone().context("Failed to clone TcpStream")?;
    let mut writer = std::io::BufWriter::new(write_stream);
    let mut reader = std::io::BufReader::new(&stream);

    // --- Parse request line ---
    let mut request_line = String::new();
    reader.read_line(&mut request_line)?;
    let parts: Vec<&str> = request_line.split_whitespace().collect();

    if parts.len() < 2 {
        let _ = write!(writer, "HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n");
        return Ok(());
    }

    let method = parts[0];
    let path = parts[1];

    debug!("{} {}", method, path);

    // --- Consume remaining headers and (for POST) body ---
    let mut content_length: usize = 0;
    loop {
        let mut line = String::new();
        reader.read_line(&mut line)?;
        if line == "\r\n" || line == "\n" || line.is_empty() {
            break;
        }
        let lower = line.to_lowercase();
        if lower.starts_with("content-length:") {
            content_length = line
                .split(':')
                .nth(1)
                .unwrap_or("")
                .trim()
                .parse::<usize>()
                .unwrap_or(0);
        }
    }

    let body_bytes = if content_length > 0 {
        let mut buf = vec![0u8; content_length];
        std::io::Read::read_exact(&mut reader, &mut buf)?;
        buf
    } else {
        Vec::new()
    };

    // --- Route ---
    let response_body = match (method, path) {
        // ------------------------------------------------------------------ //
        // GET /api/yawl/health
        // ------------------------------------------------------------------ //
        ("GET", "/api/yawl/health") => {
            let healthy = connector.get_health().await.unwrap_or(false);
            let status = if healthy { "up" } else { "down" };
            info!("GET /api/yawl/health → {}", status);
            json_200(&serde_json::json!({ "status": status }))
        }

        // ------------------------------------------------------------------ //
        // GET /api/yawl/cases
        // ------------------------------------------------------------------ //
        ("GET", "/api/yawl/cases") => match connector.list_cases().await {
            Ok(cases) => {
                info!("GET /api/yawl/cases → {} case(s)", cases.len());
                let arr: Vec<serde_json::Value> = cases
                    .iter()
                    .map(|c| serde_json::json!({ "case_id": c.case_id, "spec_id": c.spec_id }))
                    .collect();
                json_200(&serde_json::Value::Array(arr))
            }
            Err(e) => {
                warn!("GET /api/yawl/cases error: {}", e);
                json_502(&format!("YAWL engine error: {}", e))
            }
        },

        // ------------------------------------------------------------------ //
        // POST /api/yawl/discover
        // ------------------------------------------------------------------ //
        ("POST", "/api/yawl/discover") => {
            let body_str = String::from_utf8_lossy(&body_bytes);
            let parsed: serde_json::Result<serde_json::Value> = serde_json::from_str(&body_str);

            match parsed {
                Err(e) => json_400(&format!("Invalid JSON body: {}", e)),
                Ok(val) => match val.get("spec_id").and_then(|v| v.as_str()) {
                    None => json_400("Missing required field: spec_id"),
                    Some(spec_id) => {
                        let engine = crate::process::ProcessMiningEngine::new();
                        match engine.discover_from_yawl(spec_id, connector).await {
                            Ok(result) => {
                                info!(
                                    "POST /api/yawl/discover spec_id={} → places={} transitions={}",
                                    spec_id, result.places, result.transitions
                                );
                                json_200(&serde_json::json!({
                                    "algorithm": result.algorithm,
                                    "places":      result.places,
                                    "transitions": result.transitions,
                                    "arcs":        result.arcs,
                                    "fitness":     result.fitness,
                                }))
                            }
                            Err(e) => {
                                warn!("POST /api/yawl/discover error: {}", e);
                                json_502(&format!("Discovery failed: {}", e))
                            }
                        }
                    }
                },
            }
        }

        // ------------------------------------------------------------------ //
        // 404 for everything else
        // ------------------------------------------------------------------ //
        _ => {
            format!(
                "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\nAccess-Control-Allow-Origin: *\r\n\r\nNot found. Available: GET /api/yawl/health, GET /api/yawl/cases, POST /api/yawl/discover\r\n"
            )
        }
    };

    write!(writer, "{}", response_body)?;
    Ok(())
}

// ---------------------------------------------------------------------------
// Response helpers — match the style in ontology/serve.rs
// ---------------------------------------------------------------------------

fn json_200(body: &serde_json::Value) -> String {
    let json = serde_json::to_string(body).unwrap_or_else(|_| "{}".to_string());
    format!(
        "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nAccess-Control-Allow-Origin: *\r\nContent-Length: {}\r\n\r\n{}",
        json.len(),
        json
    )
}

fn json_400(msg: &str) -> String {
    let body = serde_json::json!({ "error": msg });
    let json = serde_json::to_string(&body).unwrap_or_else(|_| "{}".to_string());
    format!(
        "HTTP/1.1 400 Bad Request\r\nContent-Type: application/json\r\nAccess-Control-Allow-Origin: *\r\nContent-Length: {}\r\n\r\n{}",
        json.len(),
        json
    )
}

fn json_502(msg: &str) -> String {
    let body = serde_json::json!({ "error": msg });
    let json = serde_json::to_string(&body).unwrap_or_else(|_| "{}".to_string());
    format!(
        "HTTP/1.1 502 Bad Gateway\r\nContent-Type: application/json\r\nAccess-Control-Allow-Origin: *\r\nContent-Length: {}\r\n\r\n{}",
        json.len(),
        json
    )
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_json_200_shape() {
        let body = json_200(&serde_json::json!({ "status": "up" }));
        assert!(body.starts_with("HTTP/1.1 200 OK"));
        assert!(body.contains("application/json"));
        assert!(body.contains("\"status\":\"up\""));
    }

    #[test]
    fn test_json_400_shape() {
        let body = json_400("bad input");
        assert!(body.starts_with("HTTP/1.1 400 Bad Request"));
        assert!(body.contains("bad input"));
    }

    #[test]
    fn test_json_502_shape() {
        let body = json_502("upstream failed");
        assert!(body.starts_with("HTTP/1.1 502 Bad Gateway"));
        assert!(body.contains("upstream failed"));
    }

    #[test]
    fn test_serve_config_defaults() {
        let cfg = YawlServeConfig::default();
        assert_eq!(cfg.host, "127.0.0.1");
        assert_eq!(cfg.port, 8090);
    }
}
