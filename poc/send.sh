#!/bin/bash
# Usage: ./poc/send.sh <remoti command>
# Example: ./poc/send.sh "T hello world"
#          ./poc/send.sh "C meta space"
#          ./poc/send.sh "C enter"
#          ./poc/send.sh focus <app_id>
#          ./poc/send.sh spawn <app>
#          ./poc/send.sh layout   (cycles Super+R to keep Claude Code visible)

set -e
mkdir -p poc/logs
LOG="poc/logs/controller.log"
CMD="$*"

log() { echo "$(date -Iseconds) [SEND] $*" | tee -a "$LOG"; }

case "$1" in
  focus)
    APP_ID="$2"
    WINDOW_ID=$(niri msg -j windows | jq -r ".[] | select(.app_id == \"$APP_ID\") | .id" | tail -1)
    if [ -z "$WINDOW_ID" ]; then
      log "ERROR: window not found for app_id=$APP_ID"
      exit 1
    fi
    niri msg action focus-window --id "$WINDOW_ID"
    log "Focused $APP_ID (window $WINDOW_ID)"
    ;;
  spawn)
    shift
    niri msg action spawn -- "$@"
    log "Spawned: $*"
    ;;
  layout)
    printf "C meta r\n" | nc -q1 localhost 8080
    log "Cycled layout (Super+R)"
    ;;
  overview)
    printf "C meta o\n" | nc -q1 localhost 8080
    log "Toggled overview (Super+O)"
    ;;
  wait)
    APP_ID="$2"
    MAX="${3:-10}"
    for i in $(seq 1 "$MAX"); do
      FOUND=$(niri msg -j windows | jq -r ".[] | select(.app_id == \"$APP_ID\") | .id" | tail -1)
      if [ -n "$FOUND" ]; then
        log "Window found for $APP_ID on attempt $i (ID: $FOUND)"
        echo "$FOUND"
        exit 0
      fi
      sleep 1
    done
    log "ERROR: $APP_ID not found after $MAX attempts"
    exit 1
    ;;
  notify)
    shift
    notify-send -u critical "Remoti Eye" "$*"
    log "Notification: $*"
    ;;
  *)
    # Send raw remoti command
    printf "%s\n" "$CMD" | nc -q1 localhost 8080
    RC=$?
    if [ $RC -ne 0 ]; then
      log "ERROR: nc failed (rc=$RC) sending: $CMD"
      exit 1
    fi
    log "Sent: $CMD"
    ;;
esac
