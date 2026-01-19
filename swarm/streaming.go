package swarm

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
)

// CreateStreamingSwarm creates a multi-agent swarm graph with streaming support.
// This is the streaming version of CreateSwarm.
//
// Returns:
//   - A StreamingStateGraph ready to be compiled with CompileStreaming()
//
// Example:
//
//	workflow, err := swarm.CreateStreamingSwarm(swarm.SwarmConfig{
//	    Agents: []swarm.Agent{
//	        {Name: "Alice", Runnable: aliceAgent, Destinations: []string{"Bob"}},
//	        {Name: "Bob", Runnable: bobAgent, Destinations: []string{"Alice"}},
//	    },
//	    DefaultActiveAgent: "Alice",
//	})
//	streamingApp, _ := workflow.CompileStreaming()
//	streamResult := streamingApp.Stream(ctx, initialState)
func CreateStreamingSwarm(config SwarmConfig) (*graph.StreamingStateGraph[SwarmState], error) {
	if len(config.Agents) == 0 {
		return nil, fmt.Errorf("agents list cannot be empty")
	}

	agentNames := make([]string, len(config.Agents))
	for i, agent := range config.Agents {
		agentNames[i] = agent.Name
	}

	// Validate default active agent
	found := false
	for _, name := range agentNames {
		if name == config.DefaultActiveAgent {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("default active agent '%s' not found in agent names %v",
			config.DefaultActiveAgent, agentNames)
	}

	// Create STREAMING state graph (key difference!)
	g := graph.NewStreamingStateGraph[SwarmState]()

	// Set entry point to default active agent
	g.SetEntryPoint(config.DefaultActiveAgent)

	// Add nodes for each agent
	for _, agent := range config.Agents {
		agentCopy := agent

		nodeFunc := func(ctx context.Context, state SwarmState) (SwarmState, error) {
			// Try typed Invoke first (returns SwarmState directly)
			if invoker, ok := agentCopy.Runnable.(interface {
				Invoke(context.Context, SwarmState) (SwarmState, error)
			}); ok {
				return invoker.Invoke(ctx, state)
			}

			// Fallback to any return type
			if invoker, ok := agentCopy.Runnable.(interface {
				Invoke(context.Context, SwarmState) (any, error)
			}); ok {
				result, err := invoker.Invoke(ctx, state)
				if err != nil {
					return state, err
				}
				if resultState, ok := result.(SwarmState); ok {
					return resultState, nil
				}
			}

			return state, nil
		}

		g.AddNode(agentCopy.Name, "", nodeFunc)
	}

	// Add edges
	for _, agent := range config.Agents {
		if len(agent.Destinations) > 0 {
			// Has destinations - add conditional edge for routing
			agentCopy := agent
			g.AddConditionalEdge(agentCopy.Name, func(ctx context.Context, state SwarmState) string {
				// If active agent changed, route to new agent
				if state.ActiveAgent != "" && state.ActiveAgent != agentCopy.Name {
					// Check if destination is valid
					for _, dest := range agentCopy.Destinations {
						if dest == state.ActiveAgent {
							return state.ActiveAgent
						}
					}
				}
				return graph.END
			})
		} else {
			// No destinations - go to END
			g.AddEdge(agent.Name, graph.END)
		}
	}

	return g, nil
}
