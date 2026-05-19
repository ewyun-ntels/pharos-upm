#!/bin/bash

# Backend Type Generator (Go)
# Wrapper script for Go-based quicktype generator

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Run the Go generator
go run "$SCRIPT_DIR/generate-types/main.go"
