#!/bin/bash
set -e

echo "=== Remoti Eye POC Launcher ==="
echo ""

mkdir -p poc/logs

# Prerequisites check
echo "Checking prerequisites..."

if ! nc -z localhost 8080 2>/dev/null; then
    echo "ERROR: remoti not running on :8080"
    echo "  Start it with: sudo ./remoti"
    exit 1
fi
echo "  [OK] remoti on :8080"

if ! which grim >/dev/null 2>&1; then
    echo "ERROR: grim not installed"
    echo "  Install with: sudo apt install grim"
    exit 1
fi
echo "  [OK] grim available"

if ! niri msg version >/dev/null 2>&1; then
    echo "ERROR: niri IPC not available"
    exit 1
fi
echo "  [OK] niri IPC"

echo ""
echo "All prerequisites met."
echo ""
echo "=== How to Run ==="
echo ""
echo "In Claude Code, spawn a team with two agents:"
echo ""
echo "  1. Use /team or TeamCreate to create a team"
echo "  2. Add a 'watcher' teammate (haiku model) with prompt: poc/watcher.prompt.md"
echo "  3. Add a 'controller' teammate (sonnet model) with prompt: poc/controller.prompt.md"
echo "  4. Send the controller: 'Execute the demo sequence from your prompt'"
echo ""
echo "Or run the dry-run first to verify infrastructure:"
echo ""
echo "  bash poc/dry-run.sh"
echo ""
echo "Logs will be in poc/logs/"
