package swarm

import (
	"testing"
)

func TestCreateHandoffTool(t *testing.T) {
	tool := CreateHandoffTool(HandoffToolConfig{
		AgentName:   "TestAgent",
		Name:        "test_tool",
		Description: "Test handoff tool",
	})

	if tool.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", tool.Name())
	}

	if tool.Description() != "Test handoff tool" {
		t.Errorf("Expected description 'Test handoff tool', got '%s'", tool.Description())
	}
}

func TestCreateHandoffToolDefaultName(t *testing.T) {
	tool := CreateHandoffTool(HandoffToolConfig{
		AgentName:   "Bob",
		Description: "Transfer to Bob",
	})

	if tool.Name() != "transfer_to_bob" {
		t.Errorf("Expected default name 'transfer_to_bob', got '%s'", tool.Name())
	}
}

func TestCreateHandoffToolDefaultDescription(t *testing.T) {
	tool := CreateHandoffTool(HandoffToolConfig{
		AgentName: "Alice",
	})

	expectedDesc := "Ask agent 'Alice' for help"
	if tool.Description() != expectedDesc {
		t.Errorf("Expected default description '%s', got '%s'", expectedDesc, tool.Description())
	}
}

func TestNormalizeAgentName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SimpleAgent", "simpleagent"},
		{"Agent With Spaces", "agent_with_spaces"},
		{"  Trim Me  ", "trim_me"},
		{"Multiple   Spaces", "multiple_spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeAgentName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeAgentName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
