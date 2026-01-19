package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-hare/langchaingo_swarm/swarm"
	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// fetchDocTool creates a tool for fetching documentation
type fetchDocTool struct{}

func (t *fetchDocTool) Name() string {
	return "fetch_doc"
}

func (t *fetchDocTool) Description() string {
	return "Fetch documentation from a URL. Returns the content of the page."
}

func (t *fetchDocTool) Call(ctx context.Context, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Limit response size
	maxLen := 5000
	content := string(body)
	if len(content) > maxLen {
		content = content[:maxLen] + "...(truncated)"
	}

	return content, nil
}

func main() {
	ctx := context.Background()

	// Initialize the LLM
	model, err := openai.New(openai.WithModel("gpt-4o"))
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// LLMS.txt for LangGraph documentation
	llmsTxt := "LangGraph:https://langchain-ai.github.io/langgraph/llms.txt"
	numURLs := 3

	// Create handoff tools
	transferToPlanner := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
		AgentName:   "planner_agent",
		Description: "Transfer to the planner_agent for clarifying questions related to the user's request.",
	})

	transferToResearcher := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
		AgentName:   "researcher_agent",
		Description: "Transfer to the researcher_agent to perform research and implement the solution to the user's request.",
	})

	// Create fetch doc tool
	fetchDoc := &fetchDocTool{}

	// Planner agent system prompt
	plannerPrompt := fmt.Sprintf(`You are a planner agent. Your job is to:
1. Understand what the user wants to accomplish
2. Break down the task into clear steps
3. Identify which LangGraph documentation to fetch from: %s
4. Suggest up to %d relevant URLs to fetch
5. Transfer to the researcher_agent when you have a clear plan

Always ask clarifying questions if the request is ambiguous.`, llmsTxt, numURLs)

	// Researcher agent system prompt
	researcherPrompt := `You are a researcher agent. Your job is to:
1. Fetch and read the documentation provided by the planner
2. Synthesize the information to answer the user's question
3. Provide code examples when relevant
4. Transfer back to planner_agent if you need more clarification

Be thorough and provide complete, working solutions.`

	// Create planner agent
	plannerGraph := graph.NewStateGraph[swarm.SwarmState]()
	plannerGraph.AddNode("process", "", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
		messages := append([]llms.MessageContent{
			llms.TextParts("system", plannerPrompt),
		}, state.Messages...)

		toolsList := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        fetchDoc.Name(),
					Description: fetchDoc.Description(),
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"url": map[string]interface{}{"type": "string"},
						},
						"required": []string{"url"},
					},
				},
			},
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        transferToResearcher.Name(),
					Description: transferToResearcher.Description(),
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		}

		response, err := model.GenerateContent(ctx, messages, llms.WithTools(toolsList))
		if err != nil {
			return state, err
		}

		aiMessage := llms.TextParts("ai", response.Choices[0].Content)
		state.Messages = append(state.Messages, aiMessage)
		return state, nil
	})
	plannerGraph.SetEntryPoint("process")
	plannerGraph.AddEdge("process", graph.END)
	plannerAgent, _ := plannerGraph.Compile()

	// Create researcher agent
	researcherGraph := graph.NewStateGraph[swarm.SwarmState]()
	researcherGraph.AddNode("process", "", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
		messages := append([]llms.MessageContent{
			llms.TextParts("system", researcherPrompt),
		}, state.Messages...)

		toolsList := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        fetchDoc.Name(),
					Description: fetchDoc.Description(),
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"url": map[string]interface{}{"type": "string"},
						},
						"required": []string{"url"},
					},
				},
			},
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        transferToPlanner.Name(),
					Description: transferToPlanner.Description(),
					Parameters: map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		}

		response, err := model.GenerateContent(ctx, messages, llms.WithTools(toolsList))
		if err != nil {
			return state, err
		}

		aiMessage := llms.TextParts("ai", response.Choices[0].Content)
		state.Messages = append(state.Messages, aiMessage)
		return state, nil
	})
	researcherGraph.SetEntryPoint("process")
	researcherGraph.AddEdge("process", graph.END)
	researcherAgent, _ := researcherGraph.Compile()

	// Create the swarm
	workflow, err := swarm.CreateSwarm(swarm.SwarmConfig{
		Agents: []swarm.Agent{
			{Name: "planner_agent", Runnable: plannerAgent, Destinations: []string{"researcher_agent"}},
			{Name: "researcher_agent", Runnable: researcherAgent, Destinations: []string{"planner_agent"}},
		},
		DefaultActiveAgent: "planner_agent",
	})
	if err != nil {
		log.Fatalf("Failed to create swarm: %v", err)
	}

	// Compile the swarm
	var app any
	if compiler, ok := workflow.(interface{ Compile() (any, error) }); ok {
		app, err = compiler.Compile()
		if err != nil {
			log.Fatalf("Failed to compile swarm: %v", err)
		}
	} else {
		log.Fatal("Workflow does not support Compile()")
	}

	// Example interaction
	fmt.Println("=== Research Assistant Swarm ===")
	fmt.Println("Planner and Researcher agents working together\n")

	state := swarm.SwarmState{
		Messages: []llms.MessageContent{
			llms.TextParts("user", "How do I create a simple ReAct agent in LangGraph?"),
		},
	}

	var result any
	if invoker, ok := app.(interface {
		Invoke(context.Context, swarm.SwarmState) (any, error)
	}); ok {
		result, err = invoker.Invoke(ctx, state)
		if err != nil {
			log.Fatalf("Failed to invoke: %v", err)
		}
	}

	// Print results
	if resultState, ok := result.(swarm.SwarmState); ok {
		fmt.Printf("\nActive Agent: %s\n", resultState.ActiveAgent)
		fmt.Printf("\nConversation History (%d messages):\n", len(resultState.Messages))
		for i, msg := range resultState.Messages {
			fmt.Printf("%d. %v\n", i+1, msg)
		}
	}
}
