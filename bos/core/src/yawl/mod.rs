//! YAWL engine connector for event log import.
//!
//! Fetches XES event logs from a running YAWL engine and feeds them into the
//! BusinessOS process mining pipeline.
//!
//! # Environment variables
//! - `YAWL_ENGINE_URL` — base URL of the YAWL engine (default: `http://localhost:8080`)
//!
//! # Example
//! ```no_run
//! # tokio_test::block_on(async {
//! use bos_core::yawl::YawlConnector;
//! use bos_core::process::ProcessMiningEngine;
//!
//! let yawl = YawlConnector::from_env();
//! let engine = ProcessMiningEngine::new();
//! let result = engine.discover_from_yawl("mySpec:1.0", &yawl).await.unwrap();
//! println!("Discovered {} places", result.places);
//! # })
//! ```

pub mod http;
pub use http::{YawlServeConfig, serve as serve_yawl_api};

use crate::gateway::GatewayError;
use reqwest::Client;
use std::time::Duration;
use tracing::{debug, info, warn};

/// Connector to a running YAWL engine for event log retrieval.
#[derive(Clone, Debug)]
pub struct YawlConnector {
    engine_url: String,
    client: Client,
}

/// Summary information about a running YAWL case.
#[derive(Debug, Clone, serde::Deserialize)]
pub struct CaseInfo {
    /// YAWL case identifier (e.g. `"1"`, `"2.3"`)
    pub case_id: String,
    /// Specification identifier the case is running under.
    pub spec_id: String,
}

impl YawlConnector {
    /// Create a connector pointing at the given engine URL.
    ///
    /// The URL should be the root of the YAWL engine, e.g. `"http://localhost:8080"`.
    pub fn new(engine_url: impl Into<String>) -> Self {
        let client = Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .expect("failed to build reqwest client");

        Self {
            engine_url: engine_url.into().trim_end_matches('/').to_string(),
            client,
        }
    }

    /// Create a connector from the `YAWL_ENGINE_URL` environment variable.
    ///
    /// Falls back to `http://localhost:8080` when the variable is unset.
    pub fn from_env() -> Self {
        let url = std::env::var("YAWL_ENGINE_URL")
            .unwrap_or_else(|_| "http://localhost:8080".to_string());
        Self::new(url)
    }

    /// Fetch an XES event log for the given specification ID.
    ///
    /// Calls `GET {engine_url}/logGateway?action=getSpecificationXESLog&specID={spec_id}` and
    /// returns the raw XES XML string.
    pub async fn fetch_xes_log(&self, spec_id: &str) -> Result<String, GatewayError> {
        let url = format!(
            "{}/logGateway?action=getSpecificationXESLog&specID={}",
            self.engine_url, spec_id
        );
        debug!("Fetching XES log from {}", url);

        let response = self
            .client
            .get(&url)
            .send()
            .await
            .map_err(|e| GatewayError::ConnectionFailed(e.to_string()))?;

        let status = response.status();
        if !status.is_success() {
            let body = response.text().await.unwrap_or_default();
            warn!("YAWL log gateway returned {}: {}", status, body);
            return Err(GatewayError::ServerError(status.as_u16(), body));
        }

        let xes_xml = response
            .text()
            .await
            .map_err(|e| GatewayError::InvalidResponse(e.to_string()))?;

        info!(
            "Fetched XES log for spec_id='{}' ({} bytes)",
            spec_id,
            xes_xml.len()
        );

        Ok(xes_xml)
    }

    /// List all currently running YAWL cases.
    ///
    /// Calls `GET {engine_url}/ia?action=getAllRunningCases` and parses the
    /// XML response into a `Vec<CaseInfo>`.  Returns an empty list when no
    /// cases are running.
    pub async fn list_cases(&self) -> Result<Vec<CaseInfo>, GatewayError> {
        let url = format!("{}/ia?action=getAllRunningCases", self.engine_url);
        debug!("Listing running cases from {}", url);

        let response = self
            .client
            .get(&url)
            .send()
            .await
            .map_err(|e| GatewayError::ConnectionFailed(e.to_string()))?;

        let status = response.status();
        if !status.is_success() {
            let body = response.text().await.unwrap_or_default();
            warn!("YAWL ia?getAllRunningCases returned {}: {}", status, body);
            return Err(GatewayError::ServerError(status.as_u16(), body));
        }

        let xml = response
            .text()
            .await
            .map_err(|e| GatewayError::InvalidResponse(e.to_string()))?;

        debug!("Parsing cases XML ({} bytes)", xml.len());
        let cases = parse_cases_xml(&xml);
        info!("Found {} running YAWL case(s)", cases.len());
        Ok(cases)
    }

    /// Check whether the YAWL engine is reachable and healthy.
    ///
    /// Calls `GET {engine_url}/health` and returns `true` on a 2xx response.
    pub async fn get_health(&self) -> Result<bool, GatewayError> {
        let url = format!("{}/health", self.engine_url);
        debug!("Checking YAWL engine health at {}", url);

        let response = self
            .client
            .get(&url)
            .send()
            .await
            .map_err(|e| GatewayError::ConnectionFailed(e.to_string()))?;

        let healthy = response.status().is_success();
        info!("YAWL engine health check: {}", if healthy { "OK" } else { "DEGRADED" });
        Ok(healthy)
    }
}

/// Parse a minimal YAWL `getAllRunningCases` XML response.
///
/// Expected envelope (YAWL 2.x):
/// ```xml
/// <cases>
///   <case id="1" specID="mySpec"/>
///   ...
/// </cases>
/// ```
/// Unknown formats are tolerated: any parseable `<case>` element is kept.
fn parse_cases_xml(xml: &str) -> Vec<CaseInfo> {
    let mut cases = Vec::new();

    // Walk through every occurrence of `<case` in the document.
    let mut remaining = xml;
    while let Some(start) = remaining.find("<case") {
        remaining = &remaining[start..];

        // Find the end of this element (either `/>` or `>`).
        let end = remaining.find('>').unwrap_or(remaining.len());
        let element = &remaining[..=end];

        let case_id = extract_attr(element, "id");
        let spec_id = extract_attr(element, "specID")
            .or_else(|| extract_attr(element, "specId"))
            .or_else(|| extract_attr(element, "spec_id"));

        if let (Some(cid), Some(sid)) = (case_id, spec_id) {
            cases.push(CaseInfo {
                case_id: cid,
                spec_id: sid,
            });
        }

        // Advance past the `<case` we just processed.
        remaining = &remaining[5.min(remaining.len())..];
    }

    cases
}

/// Extract the value of a named XML attribute from a raw element string.
///
/// Handles both `attr="value"` and `attr='value'` quoting styles.
fn extract_attr(element: &str, attr: &str) -> Option<String> {
    let search = format!("{}=", attr);
    let pos = element.find(&search)?;
    let rest = &element[pos + search.len()..];
    let quote = rest.chars().next()?;
    if quote != '"' && quote != '\'' {
        return None;
    }
    let inner = &rest[1..];
    let end = inner.find(quote)?;
    Some(inner[..end].to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_connector_new_trims_trailing_slash() {
        let c = YawlConnector::new("http://localhost:8080/");
        assert_eq!(c.engine_url, "http://localhost:8080");
    }

    #[test]
    fn test_from_env_fallback() {
        // Do not modify process env — just verify the fallback URL is used when
        // YAWL_ENGINE_URL is absent.
        if std::env::var("YAWL_ENGINE_URL").is_err() {
            let c = YawlConnector::from_env();
            assert_eq!(c.engine_url, "http://localhost:8080");
        }
    }

    #[test]
    fn test_parse_cases_xml_double_quotes() {
        let xml = r#"<cases>
            <case id="42" specID="InvoiceProcess:1.0"/>
            <case id="43" specID="OrderProcess:2.0"/>
        </cases>"#;
        let cases = parse_cases_xml(xml);
        assert_eq!(cases.len(), 2);
        assert_eq!(cases[0].case_id, "42");
        assert_eq!(cases[0].spec_id, "InvoiceProcess:1.0");
        assert_eq!(cases[1].case_id, "43");
        assert_eq!(cases[1].spec_id, "OrderProcess:2.0");
    }

    #[test]
    fn test_parse_cases_xml_single_quotes() {
        let xml = r#"<cases><case id='1' specID='TestSpec:0.1'/></cases>"#;
        let cases = parse_cases_xml(xml);
        assert_eq!(cases.len(), 1);
        assert_eq!(cases[0].case_id, "1");
        assert_eq!(cases[0].spec_id, "TestSpec:0.1");
    }

    #[test]
    fn test_parse_cases_xml_empty() {
        let cases = parse_cases_xml("<cases/>");
        assert!(cases.is_empty());
    }

    #[test]
    fn test_parse_cases_xml_alt_attr_names() {
        // YAWL versions may use specId (camel) or spec_id (snake)
        let xml = r#"<cases><case id="7" specId="Alt:1.0"/></cases>"#;
        let cases = parse_cases_xml(xml);
        assert_eq!(cases.len(), 1);
        assert_eq!(cases[0].spec_id, "Alt:1.0");
    }

    #[test]
    fn test_extract_attr_double_quotes() {
        let elem = r#"<case id="99" specID="Foo"/>"#;
        assert_eq!(extract_attr(elem, "id"), Some("99".to_string()));
        assert_eq!(extract_attr(elem, "specID"), Some("Foo".to_string()));
    }

    #[test]
    fn test_extract_attr_missing() {
        let elem = r#"<case id="1"/>"#;
        assert_eq!(extract_attr(elem, "specID"), None);
    }
}
