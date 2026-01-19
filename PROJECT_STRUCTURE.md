# Project Structure

This document describes the organization of the LangGraphGo Swarm project.

## Directory Layout

```
langgraphgo-swarm/
├── swarm/                      # Core swarm implementation
│   ├── swarm.go               # Swarm creation and routing logic
│   ├── swarm_test.go          # Tests for swarm functionality
│   ├── handoff.go             # Handoff tool implementation
│   └── handoff_test.go        # Tests for handoff tools
├── examples/                   # Example applications
│   ├── README.md              # Examples documentation
│   ├── basic/                 # Simple two-agent example
│   │   └── main.go
│   └── customer_support/      # Customer support system example
│       └── main.go
├── go.mod                     # Go module definition
├── README.md                  # Main documentation
├── LICENSE                    # MIT license
├── CONTRIBUTING.md            # Contribution guidelines
├── Makefile                   # Build and test automation
├── .gitignore                 # Git ignore rules
└── doc.go                     # Package documentation

```

## Core Packages

### `swarm` Package

The main package containing the swarm implementation.

**Files:**

1. **`swarm.go`** - Core swarm functionality
   - `SwarmState`: State structure for multi-agent systems
   - `SwarmConfig`: Configuration for swarm creation
   - `CreateSwarm()`: Main function to create a swarm
   - `AddActiveAgentRouter()`: Routing logic for active agents
   - `Agent`: Agent definition structure

2. **`handoff.go`** - Handoff tool implementation
   - `CreateHandoffTool()`: Creates tools for agent handoffs
   - `HandoffToolFunc()`: Creates handoff functions
   - `GetHandoffDestinations()`: Extracts destinations from tools
   - `ProcessHandoff()`: Processes handoff responses
   - Helper functions for handoff management

3. **`swarm_test.go`** - Swarm tests
   - Tests for swarm creation and validation
   - Tests for agent routing
   - Tests for state management
   - Integration tests

4. **`handoff_test.go`** - Handoff tool tests
   - Tests for tool creation
   - Tests for tool execution
   - Tests for destination extraction
   - Tests for handoff processing

## Examples

### `examples/basic`

A simple demonstration showing:
- Two-agent swarm (Alice and Bob)
- Basic handoff between agents
- Multi-turn conversations
- Tool integration

**Key Files:**
- `main.go`: Complete working example

### `examples/customer_support`

A realistic customer support system showing:
- Multiple specialized agents
- Domain-specific tools
- Context preservation across handoffs
- Mock data management

**Key Files:**
- `main.go`: Complete customer support implementation

## Testing

Tests are located alongside their implementation files with `_test.go` suffix.

**Test Coverage:**
- Unit tests for all core functions
- Integration tests for swarm behavior
- Mock implementations for testing

**Running Tests:**
```bash
make test          # Run all tests
make coverage      # Generate coverage report
```

## Build System

The `Makefile` provides convenient targets:

```bash
make test          # Run tests
make build         # Build examples
make clean         # Clean build artifacts
make install       # Download dependencies
make lint          # Run linters
make fmt           # Format code
make check         # Run all checks
```

## Documentation

### User Documentation
- `README.md` - Main project documentation
- `examples/README.md` - Examples guide
- `doc.go` - Package-level documentation

### Developer Documentation
- `CONTRIBUTING.md` - Contribution guidelines
- `PROJECT_STRUCTURE.md` - This file
- Inline code comments and godoc

## Dependencies

Key dependencies (see `go.mod`):
- `github.com/smallnest/langgraphgo` - Graph execution engine
- `github.com/tmc/langchaingo` - LangChain for Go

## File Naming Conventions

- `*.go` - Go source files
- `*_test.go` - Test files
- `main.go` - Executable entry points
- `doc.go` - Package documentation
- `README.md` - Markdown documentation

## Code Organization Principles

1. **Separation of Concerns**
   - Core logic in `swarm` package
   - Examples in separate directory
   - Tests alongside implementation

2. **Go Idioms**
   - Interfaces for flexibility
   - Struct-based configuration
   - Explicit error handling
   - Context propagation

3. **Modularity**
   - Small, focused functions
   - Clear package boundaries
   - Minimal dependencies

4. **Testability**
   - Unit testable components
   - Mock-friendly interfaces
   - Table-driven tests

## Adding New Features

When adding new features:

1. **Implementation**
   - Add code to appropriate file in `swarm/`
   - Follow existing patterns and conventions
   - Include godoc comments

2. **Testing**
   - Add tests in `*_test.go` file
   - Cover success and error cases
   - Use table-driven tests where appropriate

3. **Documentation**
   - Update README.md if user-facing
   - Add example if significant feature
   - Include inline documentation

4. **Examples**
   - Add example if it demonstrates new capability
   - Update examples/README.md
   - Ensure example runs correctly

## Version Control

### Git Workflow
- Feature branches for development
- Pull requests for review
- Squash merges to main

### Ignored Files (`.gitignore`)
- Build artifacts
- IDE files
- OS-specific files
- Environment files
- Temporary files

## Future Structure Considerations

As the project grows, consider:

1. **Additional Packages**
   - `swarm/middleware` - Middleware for agents
   - `swarm/persistence` - Extended persistence options
   - `swarm/tools` - Reusable tool implementations

2. **Documentation**
   - API reference documentation
   - Architecture decision records
   - Performance benchmarks

3. **Tooling**
   - CI/CD workflows
   - Automated releases
   - Benchmark tracking

## Related Projects

This project is part of the LangGraph ecosystem:

- [LangGraphGo](https://github.com/smallnest/langgraphgo) - Go implementation of LangGraph
- [LangGraph (Python)](https://github.com/langchain-ai/langgraph) - Original Python implementation
- [LangGraph Swarm (Python)](https://github.com/langchain-ai/langgraph-swarm-py) - Python swarm implementation

## Maintenance

### Regular Tasks
- Keep dependencies updated
- Run tests before commits
- Format code with `make fmt`
- Update documentation as needed

### Release Process
1. Update version in documentation
2. Run full test suite
3. Update CHANGELOG (when created)
4. Tag release in Git
5. Build and verify examples

## Questions?

For questions about project structure:
- See CONTRIBUTING.md for development guidelines
- Open an issue on GitHub
- Check existing documentation
