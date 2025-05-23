#!/bin/bash

echo "Starting watch mode test..."

# Start the watch mode in background
go run cmd/go-sentinel-cli/main.go run --optimized ./internal/cli/ -w &
WATCH_PID=$!

echo "Watch mode started with PID: $WATCH_PID"

# Wait for initial startup
sleep 5

echo "Making file change..."
echo "// Test change $(date)" >> internal/cli/test_cache.go

# Wait for change detection
sleep 3

echo "Making another file change..."
echo "// Another test change $(date)" >> internal/cli/test_cache.go

# Wait for change detection
sleep 3

echo "Stopping watch mode..."
kill $WATCH_PID

echo "Test completed." 