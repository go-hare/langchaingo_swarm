package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-hare/langchaingo_swarm/swarm"
	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Simple tool for addition
func addTool(a, b int) int {
	return a + b
}

func main() {
	ctx := context.Background()

	// Initialize the LLM
	model, err := openai.New(openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create handoff tools
	transferToBob := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
		AgentName:   "Bob",
		Description: "Transfer to Bob",
	})

	transferToAlice := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
		AgentName:   "Alice",
		Description: "Transfer to Alice, she can help with math",
	})

	// Create Alice agent - addition expert
	aliceGraph := graph.NewStateGraph[swarm.SwarmState]()
	aliceGraph.AddNode("call_model", "", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
		systemPrompt := llms.TextParts("system", "You are Alice, an addition expert.")
		messages := append([]llms.MessageContent{systemPrompt}, state.Messages...)

		response, err := model.GenerateContent(ctx, messages,
			llms.WithTools([]llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name:        "add",
						Description: "Add two numbers",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"a": map[string]interface{}{"type": "integer"},
								"b": map[string]interface{}{"type": "integer"},
							},
							"required": []string{"a", "b"},
						},
					},
				},
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name:        transferToBob.Name(),
						Description: transferToBob.Description(),
						Parameters: map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
				},
			}),
		)
		if err != nil {
			return state, err
		}

		// Add AI response to messages
		aiMessage := llms.TextParts("ai", response.Choices[0].Content)
		state.Messages = append(state.Messages, aiMessage)

		return state, nil
	})
	aliceGraph.SetEntryPoint("call_model")
	aliceGraph.AddEdge("call_model", graph.END)
	alice, _ := aliceGraph.Compile()

	// Create Bob agent - pirate speaker
	bobGraph := graph.NewStateGraph[swarm.SwarmState]()
	bobGraph.AddNode("call_model", "", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
		systemPrompt := llms.TextParts("system", "You are Bob, you speak like a pirate.")
		messages := append([]llms.MessageContent{systemPrompt}, state.Messages...)

		response, err := model.GenerateContent(ctx, messages,
			llms.WithTools([]llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name:        transferToAlice.Name(),
						Description: transferToAlice.Description(),
						Parameters: map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
				},
			}),
		)
		if err != nil {
			return state, err
		}

		// Add AI response to messages
		aiMessage := llms.TextParts("ai", response.Choices[0].Content)
		state.Messages = append(state.Messages, aiMessage)

		return state, nil
	})
	bobGraph.SetEntryPoint("call_model")
	bobGraph.AddEdge("call_model", graph.END)
	bob, _ := bobGraph.Compile()

	// Create swarm with both agents
	workflow, err := swarm.CreateSwarm(swarm.SwarmConfig{
		Agents: []swarm.Agent{
			{Name: "Alice", Runnable: alice, Destinations: []string{"Bob"}},
			{Name: "Bob", Runnable: bob, Destinations: []string{"Alice"}},
		},
		DefaultActiveAgent: "Alice",
	})
	if err != nil {
		log.Fatalf("Failed to create swarm: %v", err)
	}

	// Compile the swarm
	var app any
	if compiler, ok := workflow.(interface{ Compile() (any, error) }); ok {
		var err error
		app, err = compiler.Compile()
		if err != nil {
			log.Fatalf("Failed to compile swarm: %v", err)
		}
	} else {
		log.Fatal("Workflow does not support Compile()")
	}

	// Turn 1: Ask to speak to Bob
	fmt.Println("=== Turn 1: Speaking to Bob ===")
	state1 := swarm.SwarmState{
		Messages: []llms.MessageContent{
			llms.TextParts("user", "i'd like to speak to Bob"),
		},
	}
	var result1 any
	if invoker, ok := app.(interface {
		Invoke(context.Context, swarm.SwarmState) (any, error)
	}); ok {
		result1, err = invoker.Invoke(ctx, state1)
	}
	if err != nil {
		log.Fatalf("Turn 1 failed: %v", err)
	}
	if resultState, ok := result1.(swarm.SwarmState); ok {
		fmt.Printf("Active Agent: %s\n", resultState.ActiveAgent)
		fmt.Printf("Last Message: %s\n\n", resultState.Messages[len(resultState.Messages)-1])
	}

	// Turn 2: Ask Bob to do math (should transfer to Alice)
	fmt.Println("=== Turn 2: Asking for math ===")
	state2 := result1.(swarm.SwarmState)
	state2.Messages = append(state2.Messages, llms.TextParts("user", "what's 5 + 7?"))
	var result2 any
	if invoker, ok := app.(interface {
		Invoke(context.Context, swarm.SwarmState) (any, error)
	}); ok {
		result2, err = invoker.Invoke(ctx, state2)
	}
	if err != nil {
		log.Fatalf("Turn 2 failed: %v", err)
	}
	if resultState, ok := result2.(swarm.SwarmState); ok {
		fmt.Printf("Active Agent: %s\n", resultState.ActiveAgent)
		fmt.Printf("Last Message: %s\n", resultState.Messages[len(resultState.Messages)-1])
	}
}
