// Package langgraphgo-swarm provides a multi-agent swarm system implementation for Go.
//
// LangGraphGo Swarm is a Go port of the Python LangGraph Swarm library, enabling
// creation of swarm-style multi-agent systems where agents dynamically hand off
// control to one another based on their specializations.
//
// # Key Concepts
//
// A swarm consists of multiple specialized agents that can transfer control between
// each other. The system maintains state including conversation history and tracks
// which agent is currently active.
//
// # Basic Usage
//
//	import (
//	    "context"
//	    "github.com/smallnest/langgraphgo/graph"
//	    "github.com/tmc/langchaingo/llms/openai"
//	    "github.com/yourusername/langgraphgo-swarm/swarm"
//	)
//
//	func main() {
//	    ctx := context.Background()
//	    model, _ := openai.New()
//
//	    // Create agents
//	    agent1 := createAgent(model, "Agent 1")
//	    agent2 := createAgent(model, "Agent 2")
//
//	    // Create swarm
//	    workflow, _ := swarm.CreateSwarm(swarm.SwarmConfig{
//	        Agents: []swarm.Agent{
//	            {Name: "Agent1", Runnable: agent1, Destinations: []string{"Agent2"}},
//	            {Name: "Agent2", Runnable: agent2, Destinations: []string{"Agent1"}},
//	        },
//	        DefaultActiveAgent: "Agent1",
//	    })
//
//	    // Compile and run
//	    app, _ := workflow.Compile()
//	    result, _ := app.Invoke(ctx, swarm.SwarmState{
//	        Messages: []llms.MessageContent{
//	            llms.TextParts("user", "Hello!"),
//	        },
//	    })
//	}
//
// # Features
//
// - Multi-agent collaboration with dynamic handoffs
// - Customizable handoff tools for agent communication
// - State management with conversation history tracking
// - Integration with LangGraphGo's persistence and memory
// - Type-safe agent definitions
//
// # Agent Creation
//
// Agents are created as LangGraphGo compiled graphs:
//
//	g := graph.NewStateGraph(swarm.SwarmState{})
//	g.AddNode("process", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
//	    // Agent logic
//	    return state, nil
//	})
//	g.SetEntryPoint("process")
//	g.AddEdge("process", graph.END)
//	agent, _ := g.Compile()
//
// # Handoff Tools
//
// Create tools that allow agents to transfer control:
//
//	transferTool := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
//	    AgentName:   "TargetAgent",
//	    Description: "When to transfer to this agent",
//	})
//
// The tool can then be used by an agent to hand off control to another agent.
//
// # State Management
//
// The SwarmState tracks conversation messages and the currently active agent:
//
//	type SwarmState struct {
//	    Messages    []llms.MessageContent
//	    ActiveAgent string
//	}
//
// State is automatically managed by the swarm and updated during handoffs.
//
// # Examples
//
// See the examples directory for complete working examples:
//
// - examples/basic - Simple two-agent system
// - examples/customer_support - Multi-agent customer support system
//
// # More Information
//
// For more details, see:
// - GitHub: https://github.com/yourusername/langgraphgo-swarm
// - LangGraphGo: https://github.com/smallnest/langgraphgo
// - Documentation: https://lango.rpcx.io
package main
