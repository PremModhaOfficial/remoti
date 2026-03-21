#!/bin/bash
set -e

LOG="poc/logs/dry-run.log"
mkdir -p poc/logs

log() { echo "$(date -Iseconds) [DRY-RUN] $*" | tee -a "$LOG"; }

log "Starting dry-run sequence"

# Verify prerequisites
nc -z localhost 8080 || { log "ERROR: remoti not running on :8080"; exit 1; }
which grim >/dev/null || { log "ERROR: grim not found"; exit 1; }
niri msg version >/dev/null 2>&1 || { log "ERROR: niri IPC not available"; exit 1; }

log "Prerequisites OK"

# Step 1: Open launcher (Super+Space)
printf "C meta space\n" | nc -q1 localhost 8080
[ $? -eq 0 ] || { log "ERROR: nc failed sending C meta space"; exit 1; }
log "Sent: C meta space (open launcher)"
sleep 2

# Step 2: Type ghostty and launch
printf "T ghostty\n" | nc -q1 localhost 8080
[ $? -eq 0 ] || { log "ERROR: nc failed sending T ghostty"; exit 1; }
log "Sent: T ghostty"
sleep 0.5

printf "C enter\n" | nc -q1 localhost 8080
[ $? -eq 0 ] || { log "ERROR: nc failed sending C enter"; exit 1; }
log "Sent: C enter (launch ghostty)"

# Step 3: Wait for Ghostty to start
log "Waiting 5s for Ghostty..."
sleep 5

# Step 4: Focus Ghostty
WINDOW_ID=$(niri msg -j windows | jq -r '.[] | select(.app_id == "com.mitchellh.ghostty") | .id' | head -1)
if [ -z "$WINDOW_ID" ]; then
    log "ERROR: Could not find Ghostty window"
    niri msg -j windows | jq '.[].app_id' >> "$LOG"
    exit 1
fi
niri msg action focus-window --id "$WINDOW_ID"
log "Focused Ghostty window ID: $WINDOW_ID"
sleep 1

# Step 5: Type test message
printf "T echo Hello from dry-run!\n" | nc -q1 localhost 8080
[ $? -eq 0 ] || { log "ERROR: nc failed sending test message"; exit 1; }
log "Sent: T echo Hello from dry-run!"
sleep 0.5

printf "C enter\n" | nc -q1 localhost 8080
[ $? -eq 0 ] || { log "ERROR: nc failed sending C enter"; exit 1; }
log "Sent: C enter"
sleep 1

# Step 6: Capture screenshot
SCREENSHOT="/tmp/remoti-dry-run-$(date +%s).png"
grim "$SCREENSHOT"
log "Screenshot captured: $SCREENSHOT"

log "Dry-run complete. Check $SCREENSHOT for result."
