package swarm

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
)

// SwarmState represents the state schema for the multi-agent swarm.
// It extends MessagesState with an active_agent field to track the current agent.
type SwarmState struct {
	Messages    []llms.MessageContent `json:"messages"`
	ActiveAgent string                `json:"active_agent,omitempty"`
}

// SwarmConfig holds configuration for creating a swarm
type SwarmConfig struct {
	// Agents is a list of compiled agent graphs
	Agents []Agent
	// DefaultActiveAgent is the name of the agent to start with
	DefaultActiveAgent string
	// ContextSchema specifies the schema for the context object passed to the workflow (optional)
	// This is useful for passing additional configuration or shared data to agents
	ContextSchema interface{}
}

// Agent represents a compiled agent in the swarm
type Agent struct {
	Name     string
	Runnable any // CompiledGraph from graph.Compile()
	// Destinations are the agent names this agent can hand off to
	Destinations []string
}

// CreateSwarm creates a multi-agent swarm graph.
//
// Args:
//   - config: Configuration for the swarm including agents and default active agent
//
// Returns:
//   - A StateGraph ready to be compiled
//
// Example:
//
//	workflow, err := swarm.CreateSwarm(swarm.SwarmConfig{
//	    Agents: []swarm.Agent{
//	        {Name: "Alice", Runnable: aliceAgent, Destinations: []string{"Bob"}},
//	        {Name: "Bob", Runnable: bobAgent, Destinations: []string{"Alice"}},
//	    },
//	    DefaultActiveAgent: "Alice",
//	})
//	app, _ := workflow.Compile()
func CreateSwarm(config SwarmConfig) (any, error) {
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

	// Create state graph with SwarmState
	// Note: When using typed structs, we don't need MapSchema.
	// MapSchema is only for map[string]any state types.
	g := graph.NewStateGraph[SwarmState]()

	// Add active agent router
	if err := addActiveAgentRouter(g, agentNames, config.DefaultActiveAgent); err != nil {
		return nil, err
	}

	// Add nodes for each agent - following example pattern
	for _, agent := range config.Agents {
		agentCopy := agent // Capture loop variable

		// Define the node function following the same pattern as examples
		nodeFunc := func(ctx context.Context, state SwarmState) (SwarmState, error) {
			// Invoke the agent's runnable
			// The runnable should be a compiled graph that accepts SwarmState
			if invoker, ok := agentCopy.Runnable.(interface {
				Invoke(context.Context, SwarmState) (any, error)
			}); ok {
				result, err := invoker.Invoke(ctx, state)
				if err != nil {
					return state, err
				}

				// Update state with agent's result
				if resultState, ok := result.(SwarmState); ok {
					return resultState, nil
				}
			}

			return state, nil
		}

		// Add node with name, description (empty), and function
		g.AddNode(agent.Name, "", nodeFunc)

		// Add edges to destinations
		for _, dest := range agent.Destinations {
			g.AddEdge(agent.Name, dest)
		}
	}

	return g, nil
}

// addActiveAgentRouter adds a router that routes to the currently active agent.
//
// Args:
//   - g: The StateGraph to add the router to
//   - agentNames: List of all agent names
//   - defaultActiveAgent: The default agent to route to if none is active
//
// Returns:
//   - error if validation fails
func addActiveAgentRouter(g any, agentNames []string, defaultActiveAgent string) error {
	// Validate default active agent
	found := false
	for _, name := range agentNames {
		if name == defaultActiveAgent {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("default active agent '%s' not found in routes %v",
			defaultActiveAgent, agentNames)
	}

	// Create routing function
	routeFunc := func(state SwarmState) string {
		if state.ActiveAgent != "" {
			return state.ActiveAgent
		}
		return defaultActiveAgent
	}

	// Add conditional edges from START
	pathMap := make(map[string]string)
	for _, name := range agentNames {
		pathMap[name] = name
	}

	// Use the AddConditionalEdges method - following example pattern
	if stateGraph, ok := g.(interface {
		AddConditionalEdges(string, func(SwarmState) string, map[string]string)
	}); ok {
		stateGraph.AddConditionalEdges("__start__", routeFunc, pathMap)
	}

	return nil
}

// AddActiveAgentRouter is a standalone function to add routing to an existing graph.
// This is useful for custom graph construction.
//
// Example:
//
//	g := graph.NewStateGraph(swarm.SwarmState{})
//	g.AddNode("Alice", aliceNode)
//	g.AddNode("Bob", bobNode)
//	err := swarm.AddActiveAgentRouter(g, []string{"Alice", "Bob"}, "Alice")
func AddActiveAgentRouter(g any, agentNames []string, defaultActiveAgent string) error {
	return addActiveAgentRouter(g, agentNames, defaultActiveAgent)
}
