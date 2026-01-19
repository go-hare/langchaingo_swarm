# Examples

This directory contains example implementations of LangGraphGo Swarm.

## Basic Example

**Location:** `basic/main.go`

A simple demonstration of a two-agent swarm:
- **Alice**: An addition expert who can do math
- **Bob**: A pirate-speaking agent

Shows basic agent creation, handoff tools, and multi-turn conversations.

```bash
cd basic
export OPENAI_API_KEY=your_key_here
go run main.go
```

## Customer Support Example

**Location:** `customer_support/main.go`

A realistic customer support system demonstrating:
- Multiple specialized agents (Flight Assistant, Hotel Assistant)
- Agent handoff between different service domains
- Maintaining user context and reservations across handoffs
- Mock database for flights and hotels
- Tool integration (search and booking functions)

```bash
cd customer_support
export OPENAI_API_KEY=your_key_here
go run main.go
```

### Features Demonstrated

1. **Dynamic Agent Routing**: Automatically routes to the appropriate specialist
2. **Context Preservation**: User reservations persist across agent transitions
3. **Domain-Specific Tools**: Each agent has specialized tools for their domain
4. **Natural Handoffs**: Agents can seamlessly transfer control

## Running Examples

### Prerequisites

1. Install Go 1.21 or later
2. Set your OpenAI API key:
   ```bash
   export OPENAI_API_KEY=your_api_key_here
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

### Run an Example

```bash
cd examples/<example_name>
go run main.go
```

## Creating Your Own Example

Here's a template for creating a new agent swarm:

```go
package main

import (
    "context"
    "github.com/smallnest/langgraphgo/graph"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/yourusername/langgraphgo_swarm/swarm"
)

func main() {
    ctx := context.Background()
    model, _ := openai.New()

    // 1. Create your agents
    agent1 := createYourAgent(model)
    agent2 := createAnotherAgent(model)

    // 2. Create the swarm
    workflow, _ := swarm.CreateSwarm(swarm.SwarmConfig{
        Agents: []swarm.Agent{
            {Name: "Agent1", Runnable: agent1, Destinations: []string{"Agent2"}},
            {Name: "Agent2", Runnable: agent2, Destinations: []string{"Agent1"}},
        },
        DefaultActiveAgent: "Agent1",
    })

    // 3. Compile and run
    app, _ := workflow.Compile()
    result, _ := app.Invoke(ctx, initialState)
}
```

## Example Patterns

### Pattern 1: Sequential Specialists

Agents work in sequence, each handling a specific part of the task:
```
User → Analyst → Researcher → Writer → Reviewer
```

### Pattern 2: Hub and Spoke

A coordinator agent routes to specialists:
```
           ┌─→ Flight Agent
Coordinator ├─→ Hotel Agent
           └─→ Car Rental Agent
```

### Pattern 3: Peer Collaboration

Agents can call each other as needed:
```
Alice ⟷ Bob ⟷ Charlie
```

## Tips

1. **Clear Responsibilities**: Give each agent a clear, specific role
2. **Good Descriptions**: Write clear handoff tool descriptions so agents know when to transfer
3. **Context Management**: Include relevant context in your state schema
4. **Error Handling**: Always handle errors from agent invocations
5. **Testing**: Test individual agents before combining into a swarm

## Additional Resources

- [Main README](../README.md)
- [LangGraphGo Documentation](https://lango.rpcx.io)
- [LangGraphGo Examples](https://github.com/smallnest/langgraphgo/tree/master/examples)
