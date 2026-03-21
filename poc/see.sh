#!/bin/bash
# Usage: ./poc/see.sh [label]
# Captures screenshot, logs it, prints the path
# Example: ./poc/see.sh "after opening launcher"

set -e
mkdir -p poc/logs
LOG="poc/logs/watcher.log"
LABEL="${1:-capture}"
SLUG=$(echo "$LABEL" | tr ' ' '-' | tr '[:upper:]' '[:lower:]')
SCREENSHOT="/tmp/remoti-eye-${SLUG}-$(date +%s).png"

grim "$SCREENSHOT" 2>&1
RC=$?
if [ $RC -ne 0 ]; then
  echo "$(date -Iseconds) [SEE] ERROR: grim failed (rc=$RC)" | tee -a "$LOG"
  exit 1
fi

SIZE=$(stat -c%s "$SCREENSHOT")
echo "$(date -Iseconds) [SEE] Captured: $SCREENSHOT (${SIZE} bytes) — $LABEL" | tee -a "$LOG"
echo "$SCREENSHOT"
