#!/bin/bash

# sparql_queries_bench.sh
# Measures SPARQL query performance against Oxigraph triplestore
# Tests at different data volumes: 100, 1000, 10000 triples
# Reports: mean, p95, p99 latency for each query type

set -e

# Configuration
OXIGRAPH_URL="${OXIGRAPH_URL:-http://localhost:8890}"
ITERATIONS="${ITERATIONS:-10}"
REPORT_FILE="sparql_benchmark_report.txt"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ============================================================================
# Helper Functions
# ============================================================================

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Measure query latency
# Args: query_name, sparql_query
measure_query_latency() {
    local query_name="$1"
    local sparql_query="$2"

    local latencies=()

    for i in $(seq 1 $ITERATIONS); do
        local start=$(date +%s%N)

        # Execute SPARQL query
        local response=$(curl -s -X POST "$OXIGRAPH_URL/query" \
            -H "Content-Type: application/sparql-query" \
            -d "$sparql_query" || echo "ERROR")

        local end=$(date +%s%N)

        # Calculate latency in milliseconds
        local latency_ns=$((end - start))
        local latency_ms=$(echo "scale=2; $latency_ns / 1000000" | bc)

        latencies+=("$latency_ms")

        if [ "$response" = "ERROR" ]; then
            log_warn "Query failed on iteration $i: $query_name"
        fi
    done

    # Calculate statistics
    calculate_stats "$query_name" "${latencies[@]}"
}

# Calculate mean, p95, p99 from latency array
calculate_stats() {
    local query_name="$1"
    shift
    local latencies=("$@")

    if [ ${#latencies[@]} -eq 0 ]; then
        log_warn "No latency data for $query_name"
        return
    fi

    # Sort latencies
    IFS=$'\n' sorted=($(sort -n <<<"${latencies[*]}")); unset IFS

    # Calculate mean
    local sum=0
    for lat in "${sorted[@]}"; do
        sum=$(echo "$sum + $lat" | bc)
    done
    local mean=$(echo "scale=2; $sum / ${#sorted[@]}" | bc)

    # Calculate percentiles
    local p95_idx=$(echo "scale=0; ${#sorted[@]} * 95 / 100" | bc)
    local p95="${sorted[$p95_idx]}"

    local p99_idx=$(echo "scale=0; ${#sorted[@]} * 99 / 100" | bc)
    local p99="${sorted[$p99_idx]}"

    # Min and Max
    local min="${sorted[0]}"
    local max="${sorted[$((${#sorted[@]} - 1))]}"

    # Write to report
    {
        echo "Query: $query_name"
        echo "  Mean: ${mean}ms"
        echo "  P95: ${p95}ms"
        echo "  P99: ${p99}ms"
        echo "  Min: ${min}ms"
        echo "  Max: ${max}ms"
        echo ""
    } | tee -a "$REPORT_FILE"

    log_info "Query: $query_name | Mean: ${mean}ms | P95: ${p95}ms | P99: ${p99}ms"
}

# Create test data in Oxigraph
create_test_data() {
    local triple_count="$1"

    log_info "Creating test data with $triple_count triples..."

    local construct_query="PREFIX ex: <http://example.org/>
CONSTRUCT {
  ?s ?p ?o .
} WHERE {
  VALUES (?s ?p ?o) {
    $(for i in $(seq 1 $triple_count); do
        echo "(ex:deal_$i ex:amount $i)"
      done)
  }
}"

    local response=$(curl -s -X POST "$OXIGRAPH_URL/query" \
        -H "Content-Type: application/sparql-query" \
        -d "$construct_query")

    if [ -z "$response" ]; then
        log_warn "Failed to create test data"
        return 1
    fi

    log_info "Test data created successfully"
    return 0
}

# ============================================================================
# SPARQL Query Benchmarks
# ============================================================================

benchmark_sparql_queries() {
    log_info "Starting SPARQL query benchmarks..."

    # Clear previous report
    > "$REPORT_FILE"

    # Test at different data volumes
    for triple_count in 100 1000 10000; do
        log_info "Testing with $triple_count triples..."

        {
            echo "=============================================================================="
            echo "SPARQL Query Benchmarks: $triple_count Triples"
            echo "=============================================================================="
            echo ""
        } >> "$REPORT_FILE"

        # Create test data
        create_test_data "$triple_count"

        # Query 1: Simple SELECT - all deals
        local query1="PREFIX ex: <http://example.org/>
SELECT ?deal WHERE {
  ?deal a ex:Deal .
}
LIMIT 100"
        measure_query_latency "SELECT_All_Deals_$triple_count" "$query1"

        # Query 2: Aggregate - count deals
        local query2="PREFIX ex: <http://example.org/>
SELECT (COUNT(?deal) as ?count) WHERE {
  ?deal a ex:Deal .
}"
        measure_query_latency "SELECT_Count_Deals_$triple_count" "$query2"

        # Query 3: FILTER - deals with amount > 50000
        local query3="PREFIX ex: <http://example.org/>
SELECT ?deal ?amount WHERE {
  ?deal ex:amount ?amount .
  FILTER (?amount > 50000)
}
LIMIT 100"
        measure_query_latency "SELECT_Filter_Amount_$triple_count" "$query3"

        # Query 4: JOIN - deals and related parties
        local query4="PREFIX ex: <http://example.org/>
SELECT ?deal ?buyer ?seller WHERE {
  ?deal ex:buyer ?buyer ;
        ex:seller ?seller .
}
LIMIT 100"
        measure_query_latency "SELECT_Join_Parties_$triple_count" "$query4"

        # Query 5: OPTIONAL - deals with optional compliance status
        local query5="PREFIX ex: <http://example.org/>
SELECT ?deal ?status WHERE {
  ?deal a ex:Deal .
  OPTIONAL { ?deal ex:complianceStatus ?status . }
}
LIMIT 100"
        measure_query_latency "SELECT_Optional_Status_$triple_count" "$query5"

        # Query 6: CONSTRUCT - create RDF from query
        local query6="PREFIX ex: <http://example.org/>
CONSTRUCT {
  ?deal ex:name ?name ;
        ex:amount ?amount .
} WHERE {
  ?deal a ex:Deal ;
        ex:name ?name ;
        ex:amount ?amount .
}"
        measure_query_latency "CONSTRUCT_Deal_Data_$triple_count" "$query6"

        # Query 7: Graph pattern - transitive property (compliance chain)
        local query7="PREFIX ex: <http://example.org/>
SELECT ?source ?target WHERE {
  ?source ex:requiresCompliance ?compliance .
  ?compliance ex:impliesCompliance ?target .
}
LIMIT 100"
        measure_query_latency "SELECT_Graph_Pattern_$triple_count" "$query7"

        # Query 8: UNION - deals or opportunities
        local query8="PREFIX ex: <http://example.org/>
SELECT ?item WHERE {
  { ?item a ex:Deal . } UNION { ?item a ex:Opportunity . }
}
LIMIT 100"
        measure_query_latency "SELECT_Union_$triple_count" "$query8"

        # Query 9: ORDER BY - deals sorted by amount
        local query9="PREFIX ex: <http://example.org/>
SELECT ?deal ?amount WHERE {
  ?deal ex:amount ?amount .
}
ORDER BY DESC(?amount)
LIMIT 50"
        measure_query_latency "SELECT_Order_By_Amount_$triple_count" "$query9"

        # Query 10: GROUP BY - aggregate deals by status
        local query10="PREFIX ex: <http://example.org/>
SELECT ?status (COUNT(?deal) as ?count) (SUM(?amount) as ?total) WHERE {
  ?deal ex:status ?status ;
        ex:amount ?amount .
}
GROUP BY ?status"
        measure_query_latency "SELECT_Group_By_Status_$triple_count" "$query10"

        echo "" >> "$REPORT_FILE"
    done

    # Print summary
    log_info "Benchmark complete. Report saved to $REPORT_FILE"
}

# ============================================================================
# Main
# ============================================================================

main() {
    log_info "SPARQL Query Performance Benchmarks"
    log_info "Oxigraph URL: $OXIGRAPH_URL"
    log_info "Iterations per query: $ITERATIONS"

    # Check if Oxigraph is running
    if ! curl -s "$OXIGRAPH_URL/health" > /dev/null 2>&1; then
        log_warn "Oxigraph may not be running at $OXIGRAPH_URL"
        log_warn "To start Oxigraph: docker run --rm -p 8890:7878 ghcr.io/oxigraph/oxigraph"
    fi

    # Run benchmarks
    benchmark_sparql_queries

    # Print summary
    {
        echo ""
        echo "=============================================================================="
        echo "Summary"
        echo "=============================================================================="
        echo "Total queries benchmarked: 30 (10 query types × 3 data volumes)"
        echo "Target SLA: p95 < 1000ms for most queries"
        echo ""
    } >> "$REPORT_FILE"

    cat "$REPORT_FILE"
}

main "$@"
