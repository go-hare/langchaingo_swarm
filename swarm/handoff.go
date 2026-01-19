package swarm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

const (
	// MetadataKeyHandoffDestination is the metadata key for handoff destination
	MetadataKeyHandoffDestination = "__handoff_destination"
)

var whitespaceRe = regexp.MustCompile(`\s+`)

// normalizeAgentName normalizes an agent name to be used inside the tool name
func normalizeAgentName(agentName string) string {
	normalized := strings.TrimSpace(agentName)
	normalized = whitespaceRe.ReplaceAllString(normalized, "_")
	return strings.ToLower(normalized)
}

// HandoffToolConfig holds configuration for creating a handoff tool
type HandoffToolConfig struct {
	// AgentName is the name of the agent to handoff control to
	AgentName string
	// Name is the optional name of the tool (default: transfer_to_<agent_name>)
	Name string
	// Description is the optional description for the handoff tool
	Description string
}

// handoffTool implements the tools.Tool interface for agent handoffs
type handoffTool struct {
	name        string
	description string
	agentName   string
}

func (t *handoffTool) Name() string {
	return t.name
}

func (t *handoffTool) Description() string {
	return t.description
}

func (t *handoffTool) Call(ctx context.Context, input string) (string, error) {
	// Return a special marker that the agent node will detect and convert to Command
	// The marker format is: __HANDOFF__<agent_name>
	// The agent node should parse this and return graph.Command{Goto: agent_name, Update: ...}
	return fmt.Sprintf("__HANDOFF__%s", t.agentName), nil
}

// CreateHandoffTool creates a tool that can handoff control to the requested agent.
//
// The tool returns a marker that indicates a handoff should occur.
// The swarm system will detect this and update the active agent accordingly.
//
// Args:
//   - config: Configuration for the handoff tool
//
// Returns:
//   - A tools.Tool compatible with langchaingo that can be used in agents
//
// Example:
//
//	transferToBob := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
//	    AgentName: "Bob",
//	    Description: "Transfer to Bob for pirate speak",
//	})
func CreateHandoffTool(config HandoffToolConfig) tools.Tool {
	name := config.Name
	if name == "" {
		name = fmt.Sprintf("transfer_to_%s", normalizeAgentName(config.AgentName))
	}

	description := config.Description
	if description == "" {
		description = fmt.Sprintf("Ask agent '%s' for help", config.AgentName)
	}

	return &handoffTool{
		name:        name,
		description: description,
		agentName:   config.AgentName,
	}
}

// CreateHandoffCommand creates a Command for handing off to another agent.
// This function integrates with LangGraphGo's Command API for dynamic routing.
//
// Args:
//   - targetAgent: The name of the agent to handoff to
//   - toolCallID: The ID of the tool call that triggered the handoff (optional)
//
// Returns:
//   - A Command that updates state and routes to the target agent
//
// Example:
//
//	// In agent node after detecting handoff marker:
//	if strings.HasPrefix(toolResult, "__HANDOFF__") {
//	    targetAgent := strings.TrimPrefix(toolResult, "__HANDOFF__")
//	    return CreateHandoffCommand(targetAgent, toolCallID), nil
//	}
func CreateHandoffCommand(targetAgent, toolCallID string) *graph.Command {
	// Create tool message
	toolMessage := llms.TextParts("tool",
		fmt.Sprintf("Successfully transferred to %s", targetAgent))

	// Set tool_call_id if provided
	if toolCallID != "" {
		// Add tool call ID to message metadata if needed
		// This depends on how langchaingo handles tool message IDs
	}

	// Return Command with dynamic routing and state update
	// The "messages" field will be processed by graph.AddMessages reducer
	// which provides intelligent merging with ID-based deduplication
	return &graph.Command{
		Goto: targetAgent,
		Update: map[string]any{
			"messages":     []llms.MessageContent{toolMessage},
			"active_agent": targetAgent,
		},
	}
}

// ParseHandoffResult checks if a tool result is a handoff marker and returns the target agent.
// Returns the target agent name and true if it's a handoff, empty string and false otherwise.
//
// Example:
//
//	if targetAgent, isHandoff := ParseHandoffResult(toolResult); isHandoff {
//	    return CreateHandoffCommand(targetAgent, toolCallID), nil
//	}
func ParseHandoffResult(result string) (targetAgent string, isHandoff bool) {
	const handoffPrefix = "__HANDOFF__"
	if strings.HasPrefix(result, handoffPrefix) {
		return strings.TrimPrefix(result, handoffPrefix), true
	}
	return "", false
}

// GetHandoffDestinationsFromAgent extracts handoff destinations from a compiled agent.
// This analyzes the agent's graph to find tools with handoff metadata.
//
// Args:
//   - agent: The compiled agent graph to analyze
//   - toolNodeName: Name of the tool node to inspect (default: "tools")
//
// Returns:
//   - List of agent names that can be handed off to
//
// Example:
//
//	destinations := swarm.GetHandoffDestinationsFromAgent(aliceAgent, "tools")
//	// Returns: ["Bob", "Charlie"] if Alice has handoff tools to Bob and Charlie
//
// Note: This requires the agent graph to expose its tool nodes, which may not be
// available in all LangGraphGo versions. Returns empty list if inspection is not possible.
func GetHandoffDestinationsFromAgent(agent any, toolNodeName string) []string {
	if agent == nil {
		return []string{}
	}

	// For now, return empty list - this requires introspection of the CompiledGraph
	// In a full implementation, we would:
	// 1. Get the graph structure from agent.GetGraph()
	// 2. Find the tool node by name
	// 3. Extract tools from that node
	// 4. Call GetHandoffDestinations on those tools

	// TODO: Implement graph introspection when LangGraphGo exposes graph structure API
	return []string{}
}

// isHandoffResponse checks if a tool response indicates a handoff (private helper)
func isHandoffResponse(response string) (bool, string) {
	if strings.HasPrefix(response, "__HANDOFF__") {
		agentName := strings.TrimPrefix(response, "__HANDOFF__")
		return true, agentName
	}
	return false, ""
}

// processHandoff processes a handoff in the agent execution flow (private helper).
// This should be called after tool execution to handle handoffs.
//
// Args:
//   - state: Current swarm state
//   - toolResponse: Response from tool execution
//
// Returns:
//   - Updated state with handoff processed
//   - Boolean indicating if handoff occurred
func processHandoff(state SwarmState, toolResponse string) (SwarmState, bool) {
	if isHandoff, agentName := isHandoffResponse(toolResponse); isHandoff {
		// Add tool message
		toolMessage := llms.TextParts("tool",
			fmt.Sprintf("Successfully transferred to %s", agentName))
		state.Messages = append(state.Messages, toolMessage)
		state.ActiveAgent = agentName
		return state, true
	}
	return state, false
}
