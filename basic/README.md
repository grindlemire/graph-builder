# Graph Builder

A dependency graph execution engine in Go designed for multi-team ownership. Nodes self-register via `init()` functions, enabling decentralized development where teams can add and develop nodes without touching shared orchestration code.

## How It Works

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Registration Flow                             │
│                                                                         │
│   pkg/nodes/node1/run.go   ──┐                                          │
│   pkg/nodes/node2a/run.go  ──┼──► init() ──► register.Register()        │
│   pkg/nodes/node2b/run.go  ──┤               Global Registry            │
│   pkg/nodes/node3/run.go   ──┘                     │                    │
│                                                    ▼                    │
│                                            engine.New(registry)         │
│                                                    │                    │
│                                                    ▼                    │
│                                         Topological Sort + Execute      │
└─────────────────────────────────────────────────────────────────────────┘
```

The engine:

1. Collects all registered nodes into a dependency graph
2. Performs topological sort to determine execution levels
3. Executes nodes in parallel within each level
4. Passes results from dependencies to dependent nodes

## Project Structure

```bash
├── main.go               # Entry point: builds engine, prints graph, runs
├── nodes.go              # Import manifest: one blank import per node
├── pkg/
│   ├── engine/           # Core engine: graph resolution + parallel execution
│   ├── register/         # Global registry for node self-registration
│   └── nodes/            # Each subdirectory is one node (owned by a team)
│       ├── node1/
│       │   ├── run.go    # Node definition + init() registration
│       │   └── output.go # Typed output struct + FromDeps() helper
│       ├── node2a/
│       ├── node2b/
│       └── ...
```

## Reading the Code

### Entry Point: `main.go`

```go
e := engine.New(register.Registry())  // Build engine from all registered nodes
e.PrettyPrint()                       // Visualize the dependency graph
e.Run()                               // Execute in topological order
```

### Node Registration: `nodes.go`

This file **only contains imports**. Each import triggers that package's `init()` function:

```go
import (
    _ "github.com/grindlemire/graph-builder/basic/pkg/nodes/node1"
    _ "github.com/grindlemire/graph-builder/basic/pkg/nodes/node2a"
    // ... one line per node
)
```

### Node Package Structure

Each node package has two files:

**`run.go`** — Defines the node and registers it:

```go
const ID = "node1"

func init() {
    register.Register(engine.Node{
        ID:        ID,
        DependsOn: []string{},  // or: []string{node2a.ID, node2b.ID}
        Run:       run,
    })
}

func run(deps map[string]engine.Result) (engine.Result, error) {
    // Access dependencies via type-safe helpers
    n2a, _ := node2a.FromDeps(deps)
    
    // Return your result
    return engine.Result{ID: ID, Data: Output{Message: "done"}}, nil
}
```

**`output.go`** — Typed output struct and extraction helper:

```go
type Output struct {
    Message string
}

func FromDeps(deps map[string]engine.Result) (Output, error) {
    result, ok := deps[ID]
    if !ok {
        return Output{}, fmt.Errorf("node1 result not found")
    }
    return result.Data.(Output), nil
}
```

### Engine: `pkg/engine/engine.go`

- `topoSortLevels()` — Kahn's algorithm to group nodes into parallel execution levels
- `Run()` — Executes the graph, nodes within a level run concurrently

## Independent Node development

### Adding a New Node

1. **Create your node package:**

   ```bash
   pkg/nodes/mynode/
   ├── run.go
   └── output.go
   ```

2. **Add one import line to `nodes.go`:**

   ```go
   _ "github.com/grindlemire/graph-builder/basic/pkg/nodes/mynode"
   ```

### Scalable patterns

| Concern | Solution |
|---------|----------|
| **Merge conflicts** | Minimal—each team owns their node package; `nodes.go` is append-only |
| **Coupling** | Nodes only know about their dependencies via ID constants |
| **Testing** | Each node is testable in isolation with mocked `deps` map |
| **Type safety** | `FromDeps()` helpers provide typed access to dependency outputs |

### Dependency Declaration

Nodes declare dependencies by importing the dependency package and referencing its ID:

```go
import "github.com/grindlemire/graph-builder/basic/pkg/nodes/node2a"

func init() {
    register.Register(engine.Node{
        ID:        ID,
        DependsOn: []string{node2a.ID},  // Compile-time checked
        Run:       run,
    })
}
```

## Graph Integrity Verification

### Compile-Time Cycle Detection

Because nodes import their dependencies to reference their ID constants, the Go compiler enforces acyclic dependencies at build time. If you accidentally create a circular dependency (e.g., node A depends on B, B depends on A), the build fails:

```bash
package import cycle not allowed
```

This catches cycles before any code runs—no runtime checks needed.

### Automated Tests (`graph_test.go`)

Run `go test` to verify graph integrity:

```bash
go test -v
```

The test suite validates that there are no cycles and that dependencies are properly declared and valid. It also requires no additional touch points when developing a single node, it will automatically fail any graph dependency errors. It does this through inspecting the AST for each of the node declarations.

## Example Output

```bash
┌─────────────────────────────────────────┐
│         Dependency Graph                │
└─────────────────────────────────────────┘

  ◉ node1
    ├─ depends on: (none - root node)
    └─ required by: node2a, node2b, node2c

  ◉ node2a
    ├─ depends on: node1
    └─ required by: node3

  ◉ node3
    ├─ depends on: node2a, node2b, node2c
    └─ required by: node4

┌─────────────────────────────────────────┐
│         Execution Levels                │
└─────────────────────────────────────────┘

  Level 0:
    → node1

  Level 1 (parallel):
    → node2a
    → node2b
    → node2c

  Level 2:
    → node3

  Level 3:
    → node4
```
