package tools

import (
	"fmt"

	"github.com/azhany/codecli/internal/types"
)

// Manager manages all the available tools
type Manager struct {
	tools map[string]types.Tool
}

// NewManager creates a new tool manager
func NewManager() *Manager {
	m := &Manager{
		tools: make(map[string]types.Tool),
	}

	// Register default tools
	m.RegisterTool(NewCommand())
	m.RegisterTool(NewFile())

	return m
}

// RegisterTool adds a new tool to the manager
func (m *Manager) RegisterTool(tool types.Tool) {
	m.tools[tool.Name()] = tool
}

// GetTool returns a tool by name
func (m *Manager) GetTool(name string) (types.Tool, error) {
	tool, ok := m.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

// ListTools returns a list of all registered tools
func (m *Manager) ListTools() []types.Tool {
	var tools []types.Tool
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}
