package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"remoti/pkg/client"
	"remoti/pkg/eye"
	"remoti/pkg/eye/sense"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server with remoti's Eye for sensing and acting.
type Server struct {
	eye    *eye.Eye
	mcp    *server.MCPServer
	client *client.Client
}

// NewServer creates a remoti MCP server connected to the given remoti server address.
func NewServer(addr string) (*Server, error) {
	e, err := eye.Connect(addr)
	if err != nil {
		return nil, fmt.Errorf("mcp: connect eye: %w", err)
	}

	s := &Server{
		eye:    e,
		client: e.Act(),
	}

	s.mcp = server.NewMCPServer(
		"remoti-eye",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	s.registerTools()
	return s, nil
}

// MCPServer returns the underlying MCP server for stdio serving.
func (s *Server) MCPServer() *server.MCPServer { return s.mcp }

// Close releases all resources.
func (s *Server) Close() error { return s.eye.Close() }

func (s *Server) registerTools() {
	// Keyboard tools
	s.mcp.AddTool(gomcp.NewTool("type",
		gomcp.WithDescription("Type text using the virtual keyboard"),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Text to type")),
	), s.handleType)

	s.mcp.AddTool(gomcp.NewTool("combo",
		gomcp.WithDescription("Press a key combination (e.g. ctrl+c, alt+tab, super+1)"),
		gomcp.WithString("keys", gomcp.Required(), gomcp.Description("Key combo separated by + (e.g. ctrl+shift+a)")),
	), s.handleCombo)

	s.mcp.AddTool(gomcp.NewTool("key",
		gomcp.WithDescription("Press a single key: enter, tab, escape, space, backspace, up, down, left, right, etc."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Key name")),
	), s.handleKey)

	// Mouse tools
	s.mcp.AddTool(gomcp.NewTool("click",
		gomcp.WithDescription("Click at coordinates (x, y). Left click by default."),
		gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("X coordinate")),
		gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Y coordinate")),
		gomcp.WithString("button", gomcp.Description("Button: left (default), right, middle")),
	), s.handleClick)

	s.mcp.AddTool(gomcp.NewTool("double_click",
		gomcp.WithDescription("Double-click at coordinates"),
		gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("X coordinate")),
		gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Y coordinate")),
	), s.handleDoubleClick)

	s.mcp.AddTool(gomcp.NewTool("move",
		gomcp.WithDescription("Move mouse cursor to coordinates"),
		gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("X coordinate")),
		gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Y coordinate")),
	), s.handleMove)

	s.mcp.AddTool(gomcp.NewTool("scroll",
		gomcp.WithDescription("Scroll the mouse wheel"),
		gomcp.WithNumber("dx", gomcp.Description("Horizontal scroll amount (default 0)")),
		gomcp.WithNumber("dy", gomcp.Required(), gomcp.Description("Vertical scroll amount (negative=up, positive=down)")),
	), s.handleScroll)

	s.mcp.AddTool(gomcp.NewTool("drag",
		gomcp.WithDescription("Drag from (x1,y1) to (x2,y2)"),
		gomcp.WithNumber("x1", gomcp.Required(), gomcp.Description("Start X")),
		gomcp.WithNumber("y1", gomcp.Required(), gomcp.Description("Start Y")),
		gomcp.WithNumber("x2", gomcp.Required(), gomcp.Description("End X")),
		gomcp.WithNumber("y2", gomcp.Required(), gomcp.Description("End Y")),
	), s.handleDrag)

	// Smart sensing tools (AT-SPI + Niri)
	s.mcp.AddTool(gomcp.NewTool("find",
		gomcp.WithDescription("Find UI elements by name, role, or app. Returns element names, roles, bounds, and source."),
		gomcp.WithString("name", gomcp.Description("Element name or label to search for (substring match)")),
		gomcp.WithString("role", gomcp.Description("Filter by role: button, text, menu, menu_item, window, label, check_box, combo_box, slider, tab, link")),
		gomcp.WithString("app", gomcp.Description("Filter by application ID")),
	), s.handleFind)

	s.mcp.AddTool(gomcp.NewTool("find_and_click",
		gomcp.WithDescription("Find a UI element by name and click its center. Uses AT-SPI accessibility tree."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Element name or label to find and click")),
	), s.handleFindAndClick)

	s.mcp.AddTool(gomcp.NewTool("find_and_type",
		gomcp.WithDescription("Find a UI element by name, click it to focus, then type text into it"),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Element name or label to find")),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Text to type into the element")),
	), s.handleFindAndType)

	// Window tools
	s.mcp.AddTool(gomcp.NewTool("windows",
		gomcp.WithDescription("List all open windows with their IDs, titles, and app IDs"),
	), s.handleWindows)

	// Window management
	s.mcp.AddTool(gomcp.NewTool("focus",
		gomcp.WithDescription("Focus a window by app ID or title substring. Use before typing/clicking in that window."),
		gomcp.WithString("app", gomcp.Description("App ID substring (e.g. 'zen', 'wezterm', 'ghostty')")),
		gomcp.WithString("title", gomcp.Description("Window title substring")),
	), s.handleFocus)

	s.mcp.AddTool(gomcp.NewTool("browse",
		gomcp.WithDescription("Focus the browser, open a new tab, and navigate to a URL. Handles focus switching automatically."),
		gomcp.WithString("url", gomcp.Required(), gomcp.Description("URL to navigate to")),
	), s.handleBrowse)

	// Wait / polling
	s.mcp.AddTool(gomcp.NewTool("wait_for",
		gomcp.WithDescription("Wait for a window/text condition. Polls until condition met or timeout."),
		gomcp.WithString("window_title", gomcp.Description("Wait until a window with this title substring exists")),
		gomcp.WithString("window_app", gomcp.Description("Wait until a window with this app ID exists")),
		gomcp.WithString("tmux_text", gomcp.Description("Wait until this text appears in the active tmux pane")),
		gomcp.WithString("tmux_target", gomcp.Description("Tmux target pane (default: current)")),
		gomcp.WithNumber("timeout", gomcp.Description("Timeout in seconds (default 10)")),
	), s.handleWaitFor)

	s.mcp.AddTool(gomcp.NewTool("tmux_capture",
		gomcp.WithDescription("Capture the text content of a tmux pane. Fast alternative to screenshots for terminal verification."),
		gomcp.WithString("target", gomcp.Description("Tmux target (session:window.pane). Default: active pane")),
		gomcp.WithNumber("lines", gomcp.Description("Number of history lines to capture (default 50)")),
	), s.handleTmuxCapture)

	s.mcp.AddTool(gomcp.NewTool("tmux_send",
		gomcp.WithDescription("Send keys or text to a tmux pane. Works regardless of window focus — uses PTY write, not keyboard injection."),
		gomcp.WithString("text", gomcp.Description("Text to send to the pane")),
		gomcp.WithString("keys", gomcp.Description("Special keys like Enter, Tab, Escape, C-c, C-d, Up, Down")),
		gomcp.WithString("target", gomcp.Description("Tmux target (session:window.pane). Default: active pane")),
	), s.handleTmuxSend)

	s.mcp.AddTool(gomcp.NewTool("tmux_list",
		gomcp.WithDescription("List tmux sessions, windows, and panes with their content preview"),
	), s.handleTmuxList)

	// Niri compositor tools
	s.mcp.AddTool(gomcp.NewTool("niri_action",
		gomcp.WithDescription("Execute a niri compositor action: switch workspace, move window, toggle fullscreen, etc."),
		gomcp.WithString("action", gomcp.Required(), gomcp.Description("Action: focus-workspace-N, move-window-to-workspace-N, fullscreen, maximize, close-window, switch-preset-column-width, screenshot")),
	), s.handleNiriAction)

	s.mcp.AddTool(gomcp.NewTool("spawn",
		gomcp.WithDescription("Launch an application by command name"),
		gomcp.WithString("command", gomcp.Required(), gomcp.Description("Command to execute (e.g. 'ghostty', 'zen-browser', 'nautilus')")),
	), s.handleSpawn)

	// Screenshot
	s.mcp.AddTool(gomcp.NewTool("screenshot",
		gomcp.WithDescription("Take a screenshot of the screen. Returns the image for visual analysis."),
		gomcp.WithNumber("scale", gomcp.Description("Scale factor 0.1-1.0 (default 0.5 for speed)")),
	), s.handleScreenshot)

	s.mcp.AddTool(gomcp.NewTool("screenshot_region",
		gomcp.WithDescription("Take a screenshot of a specific screen region"),
		gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("X coordinate")),
		gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Y coordinate")),
		gomcp.WithNumber("width", gomcp.Required(), gomcp.Description("Width")),
		gomcp.WithNumber("height", gomcp.Required(), gomcp.Description("Height")),
	), s.handleScreenshotRegion)

	// Utility
	s.mcp.AddTool(gomcp.NewTool("ping",
		gomcp.WithDescription("Check if the remoti input server is alive"),
	), s.handlePing)
}

// Keyboard handlers

func (s *Server) handleType(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	text, err := req.RequireString("text")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	if err := s.client.Type(text); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("type failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("typed %d characters", len(text))), nil
}

func (s *Server) handleCombo(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	keys, err := req.RequireString("keys")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	parts := splitCombo(keys)
	if err := s.client.Combo(parts...); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("combo failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("pressed %s", keys)), nil
}

func (s *Server) handleKey(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	if err := s.client.Combo(name); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("key failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("pressed %s", name)), nil
}

// Mouse handlers

func (s *Server) handleClick(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	x, y, err := requireCoords(req)
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	button := req.GetString("button", "left")
	switch button {
	case "right":
		err = s.client.RightClick(x, y)
	case "middle":
		err = s.client.MiddleClick(x, y)
	default:
		err = s.client.Click(x, y)
	}
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("click failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("%s clicked at (%d, %d)", button, x, y)), nil
}

func (s *Server) handleDoubleClick(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	x, y, err := requireCoords(req)
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	if err := s.client.DoubleClick(x, y); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("double_click failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("double-clicked at (%d, %d)", x, y)), nil
}

func (s *Server) handleMove(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	x, y, err := requireCoords(req)
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	if err := s.client.MoveTo(x, y); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("move failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("moved to (%d, %d)", x, y)), nil
}

func (s *Server) handleScroll(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	dx := int32(req.GetFloat("dx", 0))
	dy := int32(req.GetFloat("dy", 0))
	if err := s.client.Scroll(dx, dy); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("scroll failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("scrolled (%d, %d)", dx, dy)), nil
}

func (s *Server) handleDrag(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	x1 := int32(req.GetFloat("x1", 0))
	y1 := int32(req.GetFloat("y1", 0))
	x2 := int32(req.GetFloat("x2", 0))
	y2 := int32(req.GetFloat("y2", 0))
	if err := s.client.Drag(x1, y1, x2, y2); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("drag failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("dragged from (%d,%d) to (%d,%d)", x1, y1, x2, y2)), nil
}

// Sensing handlers

func (s *Server) handleFind(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	q := sense.Query{
		Name:  req.GetString("name", ""),
		Role:  sense.Role(req.GetString("role", "")),
		AppID: req.GetString("app", ""),
	}
	if q.Name == "" && q.Role == "" && q.AppID == "" {
		return gomcp.NewToolResultError("at least one of name, role, or app is required"), nil
	}

	elems, err := s.eye.Find(ctx, q)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("find failed: %v", err)), nil
	}
	if len(elems) == 0 {
		return gomcp.NewToolResultText("no elements found"), nil
	}

	result := fmt.Sprintf("found %d element(s):\n", len(elems))
	for i, e := range elems {
		b := e.Bounds()
		result += fmt.Sprintf("  [%d] name=%q role=%s app=%s bounds=(%d,%d %dx%d) source=%s\n",
			i, e.Name(), e.Role(), e.AppID(), b.X, b.Y, b.Width, b.Height, e.Source())
	}
	return gomcp.NewToolResultText(result), nil
}

func (s *Server) handleFindAndClick(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	if err := s.eye.FindAndClick(ctx, name); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("find_and_click failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("clicked element %q", name)), nil
}

func (s *Server) handleFindAndType(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	text, err := req.RequireString("text")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	if err := s.eye.FindAndType(ctx, name, text); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("find_and_type failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("typed %q into element %q", text, name)), nil
}

// Window handlers

func (s *Server) handleWindows(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	elems, err := s.eye.Find(ctx, sense.Query{Role: sense.RoleWindow})
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("windows failed: %v", err)), nil
	}
	if len(elems) == 0 {
		return gomcp.NewToolResultText("no windows found"), nil
	}

	result := fmt.Sprintf("%d window(s):\n", len(elems))
	for i, e := range elems {
		b := e.Bounds()
		result += fmt.Sprintf("  [%d] title=%q app=%s bounds=(%d,%d %dx%d)\n",
			i, e.Name(), e.AppID(), b.X, b.Y, b.Width, b.Height)
	}
	return gomcp.NewToolResultText(result), nil
}

// Window management handlers

func (s *Server) handleFocus(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	app := req.GetString("app", "")
	title := req.GetString("title", "")
	if app == "" && title == "" {
		return gomcp.NewToolResultError("at least one of app or title is required"), nil
	}
	windowID, name, err := focusWindowByMatch(ctx, app, title)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("focus failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("focused window %d: %s", windowID, name)), nil
}

func (s *Server) handleBrowse(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	url, err := req.RequireString("url")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	// Find and focus the browser window
	_, name, err := focusWindowByMatch(ctx, "zen", "")
	if err != nil {
		// Try firefox as fallback
		_, name, err = focusWindowByMatch(ctx, "firefox", "")
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("no browser window found: %v", err)), nil
		}
	}
	time.Sleep(200 * time.Millisecond) // wait for focus

	// Open new tab
	if err := s.client.Combo("ctrl", "t"); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("ctrl+t failed: %v", err)), nil
	}
	time.Sleep(300 * time.Millisecond) // wait for new tab

	// Type URL and navigate
	if err := s.client.Type(url); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("type url failed: %v", err)), nil
	}
	time.Sleep(100 * time.Millisecond)
	if err := s.client.Combo("enter"); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("enter failed: %v", err)), nil
	}

	return gomcp.NewToolResultText(fmt.Sprintf("navigated to %s in %s", url, name)), nil
}

// focusWindowByMatch finds a window by app/title and focuses it via niri IPC.
func focusWindowByMatch(ctx context.Context, app, title string) (int, string, error) {
	out, err := exec.CommandContext(ctx, "niri", "msg", "-j", "windows").Output()
	if err != nil {
		return 0, "", fmt.Errorf("niri windows: %w", err)
	}

	var windows []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		AppID string `json:"app_id"`
	}
	if err := json.Unmarshal(out, &windows); err != nil {
		return 0, "", err
	}

	for _, w := range windows {
		appMatch := app != "" && containsLower(w.AppID, app)
		titleMatch := title != "" && containsLower(w.Title, title)
		if appMatch || titleMatch {
			err := exec.CommandContext(ctx, "niri", "msg", "action", "focus-window", "--id", fmt.Sprintf("%d", w.ID)).Run()
			if err != nil {
				return 0, "", fmt.Errorf("focus window %d: %w", w.ID, err)
			}
			return w.ID, w.Title, nil
		}
	}
	return 0, "", fmt.Errorf("no window matching app=%q title=%q", app, title)
}

// Wait / polling handlers

func (s *Server) handleWaitFor(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	windowTitle := req.GetString("window_title", "")
	windowApp := req.GetString("window_app", "")
	tmuxText := req.GetString("tmux_text", "")
	tmuxTarget := req.GetString("tmux_target", "")
	timeoutSec := req.GetFloat("timeout", 10)

	if windowTitle == "" && windowApp == "" && tmuxText == "" {
		return gomcp.NewToolResultError("at least one of window_title, window_app, or tmux_text is required"), nil
	}

	deadline := time.Now().Add(time.Duration(timeoutSec * float64(time.Second)))
	for {
		// Check window conditions via niri IPC.
		if windowTitle != "" || windowApp != "" {
			out, err := exec.CommandContext(ctx, "niri", "msg", "-j", "windows").Output()
			if err == nil {
				var windows []struct {
					ID    int    `json:"id"`
					Title string `json:"title"`
					AppID string `json:"app_id"`
				}
				if json.Unmarshal(out, &windows) == nil {
					for _, w := range windows {
						titleOK := windowTitle == "" || containsLower(w.Title, windowTitle)
						appOK := windowApp == "" || containsLower(w.AppID, windowApp)
						if titleOK && appOK {
							return gomcp.NewToolResultText(fmt.Sprintf(
								"condition met: window id=%d title=%q app=%s", w.ID, w.Title, w.AppID,
							)), nil
						}
					}
				}
			}
		}

		// Check tmux pane content.
		if tmuxText != "" {
			target := tmuxTarget
			if target == "" {
				raw, err := exec.CommandContext(ctx, "tmux", "display-message", "-p",
					"#{session_name}:#{window_index}.#{pane_index}").Output()
				if err == nil {
					target = strings.TrimSpace(string(raw))
				}
			}
			args := []string{"capture-pane", "-p"}
			if target != "" {
				args = append(args, "-t", target)
			}
			out, err := exec.CommandContext(ctx, "tmux", args...).Output()
			if err == nil && strings.Contains(string(out), tmuxText) {
				return gomcp.NewToolResultText(fmt.Sprintf(
					"condition met: text %q found in tmux pane %s", tmuxText, target,
				)), nil
			}
		}

		if time.Now().After(deadline) {
			return gomcp.NewToolResultError(fmt.Sprintf(
				"timeout after %.0fs: condition not met (window_title=%q window_app=%q tmux_text=%q)",
				timeoutSec, windowTitle, windowApp, tmuxText,
			)), nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (s *Server) handleTmuxCapture(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	target := req.GetString("target", "")
	lines := int(req.GetFloat("lines", 50))

	if target == "" {
		raw, err := exec.CommandContext(ctx, "tmux", "display-message", "-p",
			"#{session_name}:#{window_index}.#{pane_index}").Output()
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("tmux display-message failed: %v", err)), nil
		}
		target = strings.TrimSpace(string(raw))
	}

	args := []string{"capture-pane", "-p", "-t", target, "-S", fmt.Sprintf("-%d", lines)}
	out, err := exec.CommandContext(ctx, "tmux", args...).Output()
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("tmux capture-pane failed: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(out)), nil
}

// Niri compositor handlers

func (s *Server) handleNiriAction(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	action, err := req.RequireString("action")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	// Parse actions with a numeric suffix: focus-workspace-1, move-window-to-workspace-2, etc.
	var args []string
	switch {
	case strings.HasPrefix(action, "focus-workspace-"):
		n := strings.TrimPrefix(action, "focus-workspace-")
		args = []string{"msg", "action", "focus-workspace", n}
	case strings.HasPrefix(action, "move-window-to-workspace-"):
		n := strings.TrimPrefix(action, "move-window-to-workspace-")
		args = []string{"msg", "action", "move-window-to-workspace", n}
	default:
		args = []string{"msg", "action", action}
	}

	out, err := exec.CommandContext(ctx, "niri", args...).CombinedOutput()
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("niri_action %q failed: %v: %s", action, err, strings.TrimSpace(string(out)))), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("niri action %q executed", action)), nil
}

func (s *Server) handleSpawn(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	command, err := req.RequireString("command")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	args := []string{"msg", "action", "spawn", "--", command}
	out, err := exec.CommandContext(ctx, "niri", args...).CombinedOutput()
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("spawn %q failed: %v: %s", command, err, strings.TrimSpace(string(out)))), nil
	}
	return gomcp.NewToolResultText(fmt.Sprintf("spawned %q", command)), nil
}

// Tmux send/list handlers

func (s *Server) handleTmuxSend(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	text := req.GetString("text", "")
	keys := req.GetString("keys", "")
	target := req.GetString("target", "")

	if text == "" && keys == "" {
		return gomcp.NewToolResultError("at least one of text or keys is required"), nil
	}

	var msgs []string

	if text != "" {
		args := []string{"send-keys", "-l"}
		if target != "" {
			args = append(args, "-t", target)
		}
		args = append(args, text)
		out, err := exec.CommandContext(ctx, "tmux", args...).CombinedOutput()
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("tmux send-keys text failed: %v: %s", err, out)), nil
		}
		msgs = append(msgs, fmt.Sprintf("sent text (%d chars)", len(text)))
	}

	if keys != "" {
		args := []string{"send-keys"}
		if target != "" {
			args = append(args, "-t", target)
		}
		args = append(args, keys)
		out, err := exec.CommandContext(ctx, "tmux", args...).CombinedOutput()
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("tmux send-keys keys failed: %v: %s", err, out)), nil
		}
		msgs = append(msgs, fmt.Sprintf("sent keys %q", keys))
	}

	return gomcp.NewToolResultText(strings.Join(msgs, ", ")), nil
}

func (s *Server) handleTmuxList(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	format := "#{session_name}:#{window_index}.#{pane_index} #{pane_title} #{pane_current_command} #{pane_width}x#{pane_height}"
	out, err := exec.CommandContext(ctx, "tmux", "list-panes", "-a", "-F", format).Output()
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("tmux list-panes failed: %v", err)), nil
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return gomcp.NewToolResultText("no tmux panes found"), nil
	}
	return gomcp.NewToolResultText(result), nil
}

// Screenshot handlers

func (s *Server) handleScreenshot(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	scale := req.GetFloat("scale", 0.5)
	cmd := exec.CommandContext(ctx, "grim", "-s", fmt.Sprintf("%.2f", scale), "-t", "png", "-l", "1", "-")
	pngBytes, err := cmd.Output()
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("screenshot failed: %v", err)), nil
	}
	b64 := base64.StdEncoding.EncodeToString(pngBytes)
	return gomcp.NewToolResultImage(
		fmt.Sprintf("screenshot (scale=%.1f, %d bytes)", scale, len(pngBytes)),
		b64,
		"image/png",
	), nil
}

func (s *Server) handleScreenshotRegion(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	x := int(req.GetFloat("x", 0))
	y := int(req.GetFloat("y", 0))
	w := int(req.GetFloat("width", 0))
	h := int(req.GetFloat("height", 0))
	geom := fmt.Sprintf("%d,%d %dx%d", x, y, w, h)
	cmd := exec.CommandContext(ctx, "grim", "-g", geom, "-t", "png", "-l", "1", "-")
	pngBytes, err := cmd.Output()
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("screenshot_region failed: %v", err)), nil
	}
	b64 := base64.StdEncoding.EncodeToString(pngBytes)
	return gomcp.NewToolResultImage(
		fmt.Sprintf("region (%s, %d bytes)", geom, len(pngBytes)),
		b64,
		"image/png",
	), nil
}

// Utility handlers

func (s *Server) handlePing(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	if err := s.client.Ping(); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("ping failed: %v", err)), nil
	}
	log.Println("remoti-mcp: ping ok")
	return gomcp.NewToolResultText("pong"), nil
}
