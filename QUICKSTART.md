# Quick Start Guide

Get started with LangGraphGo Swarm in 5 minutes!

## 📋 Prerequisites

Before you begin, ensure you have:

- **Go 1.21+** installed ([Download](https://golang.org/dl/))
- **Git** for version control
- **OpenAI API Key** ([Get one](https://platform.openai.com/api-keys))

## 🚀 Installation

### Step 1: Create a New Project

```bash
mkdir my-swarm-project
cd my-swarm-project
go mod init my-swarm-project
```

### Step 2: Install LangGraphGo Swarm

```bash
go get github.com/yourusername/langgraphgo-swarm
go get github.com/smallnest/langgraphgo
go get github.com/tmc/langchaingo
```

### Step 3: Set Your API Key

```bash
export OPENAI_API_KEY=your_api_key_here
```

## 🎯 Your First Swarm

Create a file named `main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/yourusername/langgraphgo-swarm/swarm"
)

func main() {
	ctx := context.Background()

	// 1. Create the LLM
	model, err := openai.New(openai.WithModel("gpt-4"))
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create handoff tools
	transferToBob := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
		AgentName:   "Bob",
		Description: "Transfer to Bob, the pirate expert",
	})

	transferToAlice := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
		AgentName:   "Alice",
		Description: "Transfer to Alice, the math expert",
	})

	// 3. Create Alice (math expert)
	alice := createAgent(model, "You are Alice, a helpful math expert.", transferToBob)

	// 4. Create Bob (pirate speaker)
	bob := createAgent(model, "You are Bob. You speak like a pirate!", transferToAlice)

	// 5. Create the swarm
	workflow, err := swarm.CreateSwarm(swarm.SwarmConfig{
		Agents: []swarm.Agent{
			{Name: "Alice", Runnable: alice, Destinations: []string{"Bob"}},
			{Name: "Bob", Runnable: bob, Destinations: []string{"Alice"}},
		},
		DefaultActiveAgent: "Alice",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 6. Compile the swarm
	app, err := workflow.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// 7. Run it!
	state := swarm.SwarmState{
		Messages: []llms.MessageContent{
			llms.TextParts("user", "Hello! Can I speak to Bob?"),
		},
	}

	result, err := app.Invoke(ctx, state)
	if err != nil {
		log.Fatal(err)
	}

	// 8. Print the result
	if resultState, ok := result.(swarm.SwarmState); ok {
		fmt.Printf("Active Agent: %s\n", resultState.ActiveAgent)
		lastMsg := resultState.Messages[len(resultState.Messages)-1]
		fmt.Printf("Response: %v\n", lastMsg)
	}
}

// Helper function to create an agent
func createAgent(model llms.Model, systemPrompt string, handoffTool tools.Tool) *graph.CompiledGraph {
	g := graph.NewStateGraph(swarm.SwarmState{})

	g.AddNode("process", func(ctx context.Context, state swarm.SwarmState) (swarm.SwarmState, error) {
		messages := append([]llms.MessageContent{
			llms.TextParts("system", systemPrompt),
		}, state.Messages...)

		response, err := model.GenerateContent(ctx, messages,
			llms.WithTools([]llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name:        handoffTool.Name,
						Description: handoffTool.Description,
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

		aiMessage := llms.TextParts("ai", response.Choices[0].Content)
		state.Messages = append(state.Messages, aiMessage)
		return state, nil
	})

	g.SetEntryPoint("process")
	g.AddEdge("process", graph.END)
	compiled, _ := g.Compile()
	return compiled
}
```

### Run Your Swarm

```bash
go run main.go
```

## 🎓 Understanding the Code

Let's break down what we just did:

### 1. Import Required Packages

```go
import (
    "github.com/smallnest/langgraphgo/graph"     // Graph building
    "github.com/tmc/langchaingo/llms/openai"     // LLM integration
    "github.com/yourusername/langgraphgo-swarm/swarm" // Swarm functionality
)
```

### 2. Create Handoff Tools

Handoff tools let agents transfer control to each other:

```go
transferToBob := swarm.CreateHandoffTool(swarm.HandoffToolConfig{
    AgentName:   "Bob",
    Description: "Transfer to Bob, the pirate expert",
})
```

### 3. Create Agents

Each agent is a compiled LangGraph:

```go
alice := createAgent(model, "You are Alice, a math expert.", transferToBob)
```

### 4. Create the Swarm

Combine agents into a swarm:

```go
workflow, _ := swarm.CreateSwarm(swarm.SwarmConfig{
    Agents: []swarm.Agent{
        {Name: "Alice", Runnable: alice, Destinations: []string{"Bob"}},
        {Name: "Bob", Runnable: bob, Destinations: []string{"Alice"}},
    },
    DefaultActiveAgent: "Alice",
})
```

### 5. Compile and Invoke

```go
app, _ := workflow.Compile()
result, _ := app.Invoke(ctx, initialState)
```

## 🎯 Common Patterns

### Pattern 1: Adding Memory

Add persistence with checkpointing:

```go
import "github.com/smallnest/langgraphgo/memory"

checkpointer := memory.NewMemorySaver()
app, _ := workflow.Compile(graph.WithCheckpointer(checkpointer))

config := &graph.RunnableConfig{
    Configurable: map[string]interface{}{
        "thread_id": "user_123",
    },
}

result, _ := app.InvokeWithConfig(ctx, state, config)
```

### Pattern 2: Multiple Turns

Continue the conversation:

```go
// Turn 1
result1, _ := app.Invoke(ctx, state1)

// Turn 2 - continue from previous state
state2 := result1.(swarm.SwarmState)
state2.Messages = append(state2.Messages, 
    llms.TextParts("user", "Next question"))
result2, _ := app.Invoke(ctx, state2)
```

### Pattern 3: Custom Tools

Add domain-specific tools:

```go
func searchTool(query string) string {
    // Your search logic
    return "Search results..."
}

// Add to agent's tools
tools := []llms.Tool{
    {
        Type: "function",
        Function: &llms.FunctionDefinition{
            Name:        "search",
            Description: "Search for information",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "query": map[string]interface{}{"type": "string"},
                },
            },
        },
    },
}
```

## 🐛 Troubleshooting

### Error: "API key not found"

Make sure your OpenAI API key is set:
```bash
export OPENAI_API_KEY=your_key_here
```

### Error: "module not found"

Run:
```bash
go mod tidy
```

### Agent not responding correctly

Check:
1. System prompts are clear
2. Handoff tool descriptions are descriptive
3. API key is valid and has credits

## 📚 Next Steps

Now that you have a basic swarm running:

1. **Explore Examples**: Check out `examples/` directory
   - `examples/basic` - Simple two-agent system
   - `examples/customer_support` - Full customer support system

2. **Read Documentation**:
   - [README.md](README.md) - Full documentation
   - [API Reference](doc.go) - Detailed API docs
   - [Contributing Guide](CONTRIBUTING.md) - How to contribute

3. **Try Advanced Features**:
   - Add memory/persistence
   - Create custom state schemas
   - Build multi-agent systems
   - Add human-in-the-loop

4. **Build Something Cool**:
   - Customer support bot
   - Research assistant
   - Task automation system
   - Multi-expert consultation system

## 🎉 Tips for Success

1. **Clear Agent Roles**: Give each agent a specific, clear purpose
2. **Good Descriptions**: Write detailed handoff tool descriptions
3. **Test Individually**: Test each agent before combining
4. **Start Simple**: Begin with 2-3 agents, then expand
5. **Monitor State**: Log state changes to understand flow

## 📞 Getting Help

- **Issues**: [GitHub Issues](https://github.com/yourusername/langgraphgo-swarm/issues)
- **Examples**: Check the `examples/` directory
- **Documentation**: Read the full README.md
- **Community**: Join discussions on GitHub

## 🚀 Ready to Build?

You now have everything you need to create powerful multi-agent systems with LangGraphGo Swarm!

Start with the examples, experiment with different agent combinations, and build something amazing! 🎉
