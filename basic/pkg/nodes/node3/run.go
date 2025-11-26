package node3

import (
	"fmt"

	"github.com/grindlemire/graph-builder/basic/pkg/engine"
	"github.com/grindlemire/graph-builder/basic/pkg/nodes/node2a"
	"github.com/grindlemire/graph-builder/basic/pkg/nodes/node2b"
	"github.com/grindlemire/graph-builder/basic/pkg/nodes/node2c"
	"github.com/grindlemire/graph-builder/basic/pkg/register"
)

// ID is the unique identifier for the node. It is used to reference the node
// in the graph and to identify the node in the registry.
const ID = "node3"

// init registers the node with the registry. init is called automatically by Go
// when the package is imported. This allows us to "automatically" register the node
// with the registry at startup.
func init() {
	register.Register(engine.Node{
		ID:        ID,
		DependsOn: []string{node2a.ID, node2b.ID, node2c.ID},
		Run:       run,
	})
}

// run the node's business logic and return a result that can be used
// by other nodes in the graph. It receives outputs from its dependencies (node2a, node2b, node2c).
func run(deps map[string]engine.Result) (engine.Result, error) {
	// Extract the outputs from all dependencies using their type-safe helpers
	n2a, err := node2a.FromDeps(deps)
	if err != nil {
		return engine.Result{}, err
	}

	n2b, err := node2b.FromDeps(deps)
	if err != nil {
		return engine.Result{}, err
	}

	n2c, err := node2c.FromDeps(deps)
	if err != nil {
		return engine.Result{}, err
	}

	fmt.Printf("  → Running %s (received: %q, %q, %q)\n", ID, n2a.Message, n2b.Message, n2c.Message)

	return engine.Result{
		ID: ID,
		Data: Output{
			Message: "node3 completed - all nodes passed",
		},
	}, nil
}
