package semconv

const (
	// BosGatewayDiscoverSpan is the span name for "bos.gateway.discover".
	//
	// BOS gateway discovery — forwards event-log discovery request to pm4py-rust.
	// Kind: server
	// Stability: development
	BosGatewayDiscoverSpan = "bos.gateway.discover"

	// BosGatewayConformanceSpan is the span name for "bos.gateway.conformance".
	//
	// BOS gateway conformance — forwards conformance check to pm4py-rust.
	// Kind: server
	// Stability: development
	BosGatewayConformanceSpan = "bos.gateway.conformance"

	// BosGatewayStatisticsSpan is the span name for "bos.gateway.statistics".
	//
	// BOS gateway statistics — forwards log-statistics extraction to pm4py-rust.
	// Kind: server
	// Stability: development
	BosGatewayStatisticsSpan = "bos.gateway.statistics"

	// RdfConstructSpan is the span name for "span.rdf.construct".
	//
	// SPARQL CONSTRUCT query — produces an RDF graph from an Oxigraph triplestore.
	// Kind: client
	// Stability: development
	RdfConstructSpan = "span.rdf.construct"

	// OxigraphWriteSpan is the span name for "span.oxigraph.write".
	//
	// Oxigraph write — loads Turtle/N-Triples into Oxigraph via HTTP POST /store.
	// Kind: client
	// Stability: development
	OxigraphWriteSpan = "span.oxigraph.write"

	// OxigraphQuerySpan is the span name for "span.oxigraph.query".
	//
	// Oxigraph query — runs SPARQL against the /query endpoint.
	// Kind: client
	// Stability: development
	OxigraphQuerySpan = "span.oxigraph.query"

	// BosOntologyExecuteSpan is the span name for "bos.ontology.execute".
	//
	// bos CLI SPARQL CONSTRUCT pipeline — loads PostgreSQL rows as RDF triples
	// and writes to Oxigraph.
	// Kind: internal
	// Stability: development
	BosOntologyExecuteSpan = "bos.ontology.execute"

	// BosRdfWriteSpan is the span name for "bos.rdf.write".
	//
	// bos CLI RDF write — forwards Turtle or N-Triples to Oxigraph /store via
	// HTTP proxy.
	// Kind: client
	// Stability: development
	BosRdfWriteSpan = "bos.rdf.write"

	// BosRdfQuerySpan is the span name for "bos.rdf.query".
	//
	// bos CLI SPARQL query — proxies SELECT or CONSTRUCT to Oxigraph /query.
	// Kind: client
	// Stability: development
	BosRdfQuerySpan = "bos.rdf.query"

	// BoardL0SyncSpan is the span name for "board.l0_sync".
	//
	// Periodic L0 sync — exports BusinessOS cases/handoffs to Oxigraph via bos CLI.
	// Kind: internal
	// Stability: development
	BoardL0SyncSpan = "board.l0_sync"
)
