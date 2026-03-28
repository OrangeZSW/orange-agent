#!/bin/bash
go mod tidy 2>&1
echo "=== Build ==="
go build -o orange-agent main.go 2>&1
