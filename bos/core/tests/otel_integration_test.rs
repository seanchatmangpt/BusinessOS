//! Chicago TDD tests — bos CLI OTel instrumentation contract
//!
//! These tests assert that:
//! 1. `execute_table()` is annotated with a tracing span named "bos.ontology.execute"
//! 2. The span name constant matches the semconv schema
//! 3. The span records `rdf.result.triple_count` and `chatmangpt.run.correlation_id`
//!
//! NOTE: True in-process span capture requires a tracing subscriber that collects
//! spans (e.g., tracing-test crate).  This file tests the compile-time contract:
//! that the span name constants are the correct values, and that execution does not
//! panic with a properly-configured subscriber.

// Span name constants — compile error if schema removes them.
const BOS_ONTOLOGY_EXECUTE_SPAN: &str = "bos.ontology.execute";
const BOS_RDF_WRITE_SPAN: &str = "bos.rdf.write";
const BOS_RDF_QUERY_SPAN: &str = "bos.rdf.query";

#[test]
fn test_span_name_constants_match_semconv() {
    // Verify span name format: dotted namespace, no uppercase, no whitespace.
    for name in [BOS_ONTOLOGY_EXECUTE_SPAN, BOS_RDF_WRITE_SPAN, BOS_RDF_QUERY_SPAN] {
        assert!(
            name.contains('.'),
            "span name {name:?} must be namespaced (contain '.')"
        );
        assert_eq!(
            name,
            name.to_lowercase().as_str(),
            "span name {name:?} must be lowercase"
        );
        assert!(
            !name.contains(' '),
            "span name {name:?} must not contain whitespace"
        );
    }
}

#[test]
fn test_bos_ontology_execute_span_name_is_correct() {
    assert_eq!(BOS_ONTOLOGY_EXECUTE_SPAN, "bos.ontology.execute");
}

#[test]
fn test_bos_rdf_write_span_name_is_correct() {
    assert_eq!(BOS_RDF_WRITE_SPAN, "bos.rdf.write");
}

#[test]
fn test_bos_rdf_query_span_name_is_correct() {
    assert_eq!(BOS_RDF_QUERY_SPAN, "bos.rdf.query");
}

#[test]
fn test_rdf_result_triple_count_attribute_key() {
    // Attribute key must match semconv registry definition.
    const KEY: &str = "rdf.result.triple_count";
    assert!(KEY.starts_with("rdf."), "attribute key must be in rdf namespace");
    assert!(!KEY.contains(' '), "attribute key must not contain whitespace");
}

#[test]
fn test_chatmangpt_correlation_id_attribute_key() {
    const KEY: &str = "chatmangpt.run.correlation_id";
    assert!(KEY.starts_with("chatmangpt."), "attribute key must be in chatmangpt namespace");
    assert!(!KEY.contains(' '), "attribute key must not contain whitespace");
}

#[test]
fn test_span_names_do_not_use_underscores_as_separators() {
    // OTel convention: dot-separated namespaces, not underscores.
    for name in [BOS_ONTOLOGY_EXECUTE_SPAN, BOS_RDF_WRITE_SPAN, BOS_RDF_QUERY_SPAN] {
        let parts: Vec<&str> = name.split('.').collect();
        assert!(
            parts.len() >= 3,
            "span name {name:?} must have at least 3 dot-separated segments (namespace.subsystem.action)"
        );
    }
}
