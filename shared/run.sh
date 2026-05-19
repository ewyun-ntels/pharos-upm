#!/bin/bash

# Backend Type Generator (Go)
# Wrapper for Go-based generator

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Run the Go generator
"$SCRIPT_DIR/scripts/generate-types.sh"
