#!/usr/bin/env bash
# Seed L0 facts into Oxigraph for the Board Intelligence chain.
# Run once to bootstrap, or on reset.
#
# Usage: bash seed_l0.sh [oxigraph_url]
#
# WvdA: L0 = ground truth. Must exist before L1 materialization runs.

set -euo pipefail

OXIGRAPH_URL="${1:-http://localhost:7878}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Seeding L0 facts into ${OXIGRAPH_URL}..."

# Execute L0 CONSTRUCT and store results
curl -s -X POST \
  -H "Content-Type: application/sparql-query" \
  -H "Accept: text/turtle" \
  --data-binary @"${SCRIPT_DIR}/l0_from_businessos.sparql" \
  "${OXIGRAPH_URL}/query" | \
curl -s -X POST \
  -H "Content-Type: text/turtle" \
  "${OXIGRAPH_URL}/store?graph=http%3A%2F%2Fbusinessos.local%2Fl0"

echo "L0 facts seeded successfully."
