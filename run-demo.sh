#!/usr/bin/env bash

set -e

echo "Cleaning previous demo state..."

# Kill old processes
pkill glazeld 2>/dev/null || true
pkill glazel-agent 2>/dev/null || true
pkill redis-server 2>/dev/null || true

sleep 1

# Remove cache
rm -rf .glazel

echo "Building binaries..."

go build -o glazeld ./cmd/glazeld
go build -o glazel-agent ./cmd/glazel-agent
go build -o glazel ./cmd/glazel

echo "Demo ready."

vhs tapes/redis.tape
vhs tapes/orchestrator.tape
vhs tapes/agent.tape
vhs tapes/client.tape
