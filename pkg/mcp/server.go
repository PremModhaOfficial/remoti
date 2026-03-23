package mcp

import (
	"context"
	"fmt"
	"log"

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

// Utility handlers

func (s *Server) handlePing(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	if err := s.client.Ping(); err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("ping failed: %v", err)), nil
	}
	log.Println("remoti-mcp: ping ok")
	return gomcp.NewToolResultText("pong"), nil
}
