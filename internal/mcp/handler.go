package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	// Protocol version
	ProtocolVersion = "2024-11-05"

	// Server info
	ServerName    = "auto-wx-post-mcp"
	ServerVersion = "1.0.0"
)

// Handler handles MCP protocol communication via stdio
type Handler struct {
	server *Server
	reader *bufio.Reader
	writer *bufio.Writer
}

// NewHandler creates a new MCP handler
func NewHandler(server *Server) *Handler {
	return &Handler{
		server: server,
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
}

// Run starts the MCP server loop
func (h *Handler) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := h.handleRequest(ctx); err != nil {
				if err == io.EOF {
					return nil
				}
				// Log error but continue
				h.server.log.Error("Error handling request", "error", err)
			}
		}
	}
}

func (h *Handler) handleRequest(ctx context.Context) error {
	// Read request line
	line, err := h.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		h.sendError(nil, -32700, "Parse error", nil)
		return nil
	}

	// Handle method
	switch req.Method {
	case "initialize":
		return h.handleInitialize(req)
	case "initialized":
		// Notification, no response needed
		return nil
	case "tools/list":
		return h.handleListTools(req)
	case "tools/call":
		return h.handleCallTool(ctx, req)
	default:
		h.sendError(req.ID, -32601, "Method not found", nil)
		return nil
	}
}

func (h *Handler) handleInitialize(req JSONRPCRequest) error {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsServerCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}

	return h.sendResult(req.ID, result)
}

func (h *Handler) handleListTools(req JSONRPCRequest) error {
	tools := h.server.GetTools()
	result := ListToolsResult{
		Tools: tools,
	}

	return h.sendResult(req.ID, result)
}

func (h *Handler) handleCallTool(ctx context.Context, req JSONRPCRequest) error {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.sendError(req.ID, -32602, "Invalid params", nil)
		return nil
	}

	result, err := h.server.CallTool(ctx, params)
	if err != nil {
		h.sendError(req.ID, -32603, "Internal error", err.Error())
		return nil
	}

	return h.sendResult(req.ID, result)
}

func (h *Handler) sendResult(id interface{}, result interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return h.sendError(id, -32603, "Internal error", err.Error())
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}

	return h.writeResponse(response)
}

func (h *Handler) sendError(id interface{}, code int, message string, data interface{}) error {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	return h.writeResponse(response)
}

func (h *Handler) writeResponse(response JSONRPCResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	if _, err := h.writer.Write(data); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	if err := h.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	if err := h.writer.Flush(); err != nil {
		return fmt.Errorf("flush writer: %w", err)
	}

	return nil
}
