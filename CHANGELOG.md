# Remoti Changelog

## Session 2026-03-24 (Night)

### MCP Server — 23 Native Claude Code Tools
Built and shipped `remoti-mcp`, an MCP server that gives Claude Code native
desktop automation tools. No more `echo | nc` — direct tool calls at 55µs.

**Tools added:**
- Input: `type`, `combo`, `key`, `click`, `double_click`, `move`, `scroll`, `drag`
- Sensing: `find`, `find_and_click`, `find_and_type`, `windows`
- Window: `focus`, `browse`, `niri_action`, `spawn`
- Terminal: `tmux_send`, `tmux_list`, `tmux_capture`
- Verification: `wait_for`, `screenshot`, `screenshot_region`
- Utility: `ping`

**Key highlight:** `browse("miruro.tv")` successfully opened Zen Browser,
navigated to the anime site, and searched for JoJo — all via MCP tool calls.

### AT-SPI D-Bus Fixes
Fixed three bugs in AT-SPI accessibility tree access:
1. `ChildCount` is a D-Bus property, not a method
2. `GetChildAtIndex` returns a struct `(so)`, not two values
3. `GetExtents` returns a struct `(iiii)`, not four values

AT-SPI now walks the full accessibility tree and returns real pixel bounds.

### Client Reconnection
TCP client now auto-reconnects on broken pipe — both `send()` and `sendBatch()`.
The MCP server survives remoti input server restarts.

### Scroll Implementation
Added dedicated `uinput.Mouse` device for scroll wheel support.
TouchPad only supports absolute positioning; Mouse supports `Wheel()` API.

### TCP_NODELAY
Set on both client and server — eliminates up to 40ms Nagle delay.

### Kanata Device Exclusion
Added `linux-dev-names-exclude` for remoti virtual devices.
Home-row mods no longer eat virtual keyboard input.

### jj Snapshot Fix
Added binary artifacts to `.gitignore` — no more "Refused to snapshot" warnings.

---

## Session 2026-03-23 (Evening) — Research

### Speed Knowledgebase (10 parallel research agents)
- Input injection latency (gaming/KVM industry)
- Focus-stealing elimination (Wayland architecture)
- Event-driven verification (niri/tmux/AT-SPI/inotify)
- Fastest IPC patterns (Unix socket, MCP, shared memory)
- MCP server as native integration (55µs, 360x faster)
- GPU screen capture (PipeWire, DMA-BUF, NVENC)
- Real-time OCR (PaddleOCR+TensorRT, YOLO UI detection)
- Video surveillance frame analysis
- Browser/compositor rendering internals
- Live video transcription as event source

### Zig Comptime Differ
- Fixed 16x9 grid with XXH64: 0.45ms (benchmarked, beats quadtree)
- Zig SIMD eq-diff: 317µs for static screens
- eventfd + shared memory IPC: 1-2µs between Zig and Go daemons

### Real-Time Vision Research
- Nobody does real-time desktop vision (<100ms)
- Our event-driven architecture (AT-SPI + niri + Zig diff) is faster
  than Anthropic Computer Use, OpenAI Operator, and OmniParser V2

---

## Session 2026-03-23 (Afternoon) — Core Build

### Remoti Server
- TCP server with uinput keyboard + mouse injection
- Protocol: T/C/D/U/R (keyboard) + M C/M/R/K/D/U/S (mouse) + P (ping)
- Client library with persistent TCP, batch API, write mutex

### Eye Sensing Layer
- AT-SPI D-Bus source (accessibility tree)
- Niri IPC window source
- Cascading Finder engine (tries sources by priority)
- Element type with Click, Type, Hover, DragTo actions
- Hybrid COORD_TYPE_WINDOW + Niri offset for Wayland coordinates

### Tests
- 136 key mapping tests
- Protocol parser tests
- Client mock TCP tests

### POC Demos
- Ghostty launch + type + verify
- Zen Browser GSoC search
- Calculator via AT-SPI coords
- Neovim in tmux via sesh
