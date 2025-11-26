# Graph Builder - Server

An HTTP service that dynamically builds and executes dependency subgraphs on-demand. Extends the [basic](../basic/) engine with a `Builder` that resolves transitive dependencies at runtime, enabling per-request graph construction.

## How It Works

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Request Flow                                  │
│                                                                         │
│   pkg/nodes/node1/run.go   ──┐                                          │
│   pkg/nodes/node2a/run.go  ──┼──► init() ──► catalog.Register()         │
│   pkg/nodes/node2b/run.go  ──┤               Global Catalog             │
│   pkg/nodes/node3/run.go   ──┘                     │                    │
│                                                    ▼                    │
│                                         engine.NewBuilder(catalog)      │
│                                                    │                    │
│                                                    ▼                    │
│                              HTTP Request ──► builder.BuildFor(targets) │
│                                                    │                    │
│                                                    ▼                    │
│                                         Resolve deps + Execute subgraph │
└─────────────────────────────────────────────────────────────────────────┘
```

Key difference from `basic/`: instead of running the entire graph once, the server dynamically constructs **minimal subgraphs** per request by specifying only the terminal nodes needed.

## Project Structure

```bash
├── main.go               # HTTP server with 3 endpoints, runs demo client
├── nodes.go              # Import manifest: one blank import per node
├── pkg/
│   ├── engine/           # Core engine + Builder for dynamic graph construction
│   ├── catalog/          # Global catalog for node self-registration
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
// Create a builder from the full node catalog
builder := engine.NewBuilder(catalog.All())

// Each endpoint builds a different subgraph
mux.HandleFunc("/graph/small", handleSmallGraph(builder))   // node4 only
mux.HandleFunc("/graph/full", handleFullGraph(builder))     // node3 + all deps
mux.HandleFunc("/graph/custom", handleCustomGraph(builder)) // ?nodes=node2a,node4
```

### Dynamic Graph Construction

The `Builder.BuildFor()` method accepts terminal node IDs and automatically resolves all transitive dependencies:

```go
// Request only node4 - node1 is auto-included as a dependency
e, err := builder.BuildFor(node4.ID)

// Request multiple terminal nodes
e, err := builder.BuildFor("node2a", "node4")
```

### HTTP Endpoints

| Endpoint | Description | Example |
|----------|-------------|---------|
| `/graph/small` | Minimal graph: node1 → node4 | `GET /graph/small` |
| `/graph/full` | Full graph ending at node3 | `GET /graph/full` |
| `/graph/custom` | Custom subgraph from query params | `GET /graph/custom?nodes=node2a,node4` |

### Node Package Structure

Same as `basic/`—each node has two files:

**`run.go`** — Defines the node and registers to the catalog:

```go
const ID = "node2a"

func init() {
    catalog.Register(engine.Node{
        ID:        ID,
        DependsOn: []string{node1.ID},
        Run:       run,
    })
}

func run(deps map[string]engine.Result) (engine.Result, error) {
    n1, _ := node1.FromDeps(deps)
    return engine.Result{ID: ID, Data: Output{Message: n1.Message + " → node2a"}}, nil
}
```

**`output.go`** — Typed output struct and extraction helper (unchanged from basic).

## Builder vs Direct Engine

| Approach | Use Case |
|----------|----------|
| `engine.New(registry)` | Run full graph once (CLI tools, batch jobs) |
| `builder.BuildFor(ids...)` | Build minimal subgraph per request (servers, APIs) |

The builder pattern enables:

- **Lazy execution**: Only run nodes actually needed for a request
- **Multiple configurations**: Same catalog, different subgraphs per endpoint
- **Dynamic targets**: Accept node IDs from query params

## Running the Demo

```bash
cd server
go run .
```

The demo starts an HTTP server, runs client requests against all three endpoints, then shuts down. Example output:

```
═══════════════════════════════════════
CLIENT: Requesting Small Graph (node4 only)
        URL: http://localhost:8080/graph/small
═══════════════════════════════════════

=== /graph/small ===
┌─────────────────────────────────────┐
│         Dependency Graph            │
└─────────────────────────────────────┘

  ◉ node1
    ├─ depends on: (none - root node)
    └─ required by: node4

  ◉ node4
    ├─ depends on: node1
    └─ required by: (none - leaf node)

┌─────────────────────────────────────┐
│         Execution Levels            │
└─────────────────────────────────────┘

  Level 0:
    → node1

  Level 1:
    → node4
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

## Adding a New Node

Same process as `basic/`:

1. **Create your node package** under `pkg/nodes/mynode/`
2. **Register with catalog** in `run.go`:

   ```go
   func init() {
       catalog.Register(engine.Node{...})
   }
   ```

3. **Add import to `nodes.go`**:

   ```go
   _ "github.com/grindlemire/graph-builder/server/pkg/nodes/mynode"
   ```

The node is now available to any `BuildFor()` call that references it.
