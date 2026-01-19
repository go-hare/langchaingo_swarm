package swarm

import (
	"context"
	"fmt"
	"testing"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

// Mock agent for testing
func createMockAgent(name string, response string) any {
	g := graph.NewStateGraph[SwarmState]()

	g.AddNode("process", "", func(ctx context.Context, state SwarmState) (SwarmState, error) {
		aiMessage := llms.TextParts("ai", response)
		state.Messages = append(state.Messages, aiMessage)
		return state, nil
	})

	g.SetEntryPoint("process")
	g.AddEdge("process", graph.END)

	compiled, _ := g.Compile()
	return compiled
}

func TestCreateSwarm(t *testing.T) {
	tests := []struct {
		name        string
		config      SwarmConfig
		expectError bool
	}{
		{
			name: "valid swarm with two agents",
			config: SwarmConfig{
				Agents: []Agent{
					{Name: "Alice", Runnable: createMockAgent("Alice", "Hello from Alice"), Destinations: []string{"Bob"}},
					{Name: "Bob", Runnable: createMockAgent("Bob", "Hello from Bob"), Destinations: []string{"Alice"}},
				},
				DefaultActiveAgent: "Alice",
			},
			expectError: false,
		},
		{
			name: "empty agents list",
			config: SwarmConfig{
				Agents:             []Agent{},
				DefaultActiveAgent: "Alice",
			},
			expectError: true,
		},
		{
			name: "invalid default agent",
			config: SwarmConfig{
				Agents: []Agent{
					{Name: "Alice", Runnable: createMockAgent("Alice", "Hello"), Destinations: []string{}},
				},
				DefaultActiveAgent: "Bob",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateSwarm(tt.config)
			if (err != nil) != tt.expectError {
				t.Errorf("CreateSwarm() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestAddActiveAgentRouter(t *testing.T) {
	tests := []struct {
		name               string
		agentNames         []string
		defaultActiveAgent string
		expectError        bool
	}{
		{
			name:               "valid router",
			agentNames:         []string{"Alice", "Bob"},
			defaultActiveAgent: "Alice",
			expectError:        false,
		},
		{
			name:               "invalid default agent",
			agentNames:         []string{"Alice", "Bob"},
			defaultActiveAgent: "Charlie",
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := graph.NewStateGraph[SwarmState]()
			err := AddActiveAgentRouter(g, tt.agentNames, tt.defaultActiveAgent)
			if (err != nil) != tt.expectError {
				t.Errorf("AddActiveAgentRouter() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestSwarmStateManagement(t *testing.T) {
	ctx := context.Background()

	// Create mock agents
	alice := createMockAgent("Alice", "Alice speaking")
	bob := createMockAgent("Bob", "Bob speaking")

	// Create swarm
	workflow, err := CreateSwarm(SwarmConfig{
		Agents: []Agent{
			{Name: "Alice", Runnable: alice, Destinations: []string{"Bob"}},
			{Name: "Bob", Runnable: bob, Destinations: []string{"Alice"}},
		},
		DefaultActiveAgent: "Alice",
	})
	if err != nil {
		t.Fatalf("Failed to create swarm: %v", err)
	}

	app, err := func() (any, error) {
		if compiler, ok := workflow.(interface{ Compile() (any, error) }); ok {
			return compiler.Compile()
		}
		return nil, fmt.Errorf("workflow does not support Compile()")
	}()
	if err != nil {
		t.Fatalf("Failed to compile swarm: %v", err)
	}

	// Test initial state routing
	initialState := SwarmState{
		Messages: []llms.MessageContent{
			llms.TextParts("user", "Hello"),
		},
	}

	result, err := func() (any, error) {
		if invoker, ok := app.(interface {
			Invoke(context.Context, SwarmState) (any, error)
		}); ok {
			return invoker.Invoke(ctx, initialState)
		}
		return nil, fmt.Errorf("app does not support Invoke()")
	}()
	if err != nil {
		t.Fatalf("Failed to invoke: %v", err)
	}

	resultState, ok := result.(SwarmState)
	if !ok {
		t.Fatalf("Result is not SwarmState")
	}

	// Should start with Alice (default)
	if len(resultState.Messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(resultState.Messages))
	}

	// Test routing to specific agent
	stateWithAgent := SwarmState{
		Messages: []llms.MessageContent{
			llms.TextParts("user", "Hello Bob"),
		},
		ActiveAgent: "Bob",
	}

	result2, err := func() (any, error) {
		if invoker, ok := app.(interface {
			Invoke(context.Context, SwarmState) (any, error)
		}); ok {
			return invoker.Invoke(ctx, stateWithAgent)
		}
		return nil, fmt.Errorf("app does not support Invoke()")
	}()
	if err != nil {
		t.Fatalf("Failed to invoke with active agent: %v", err)
	}

	resultState2, ok := result2.(SwarmState)
	if !ok {
		t.Fatalf("Result is not SwarmState")
	}

	if len(resultState2.Messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(resultState2.Messages))
	}
}
