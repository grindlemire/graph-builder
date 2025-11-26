package node2c

import (
	"fmt"

	"github.com/grindlemire/graph-builder/basic/pkg/engine"
	"github.com/grindlemire/graph-builder/basic/pkg/nodes/node1"
	"github.com/grindlemire/graph-builder/basic/pkg/register"
)

// ID is the unique identifier for the node. It is used to reference the node
// in the graph and to identify the node in the registry.
const ID = "node2c"

// init registers the node with the registry. init is called automatically by Go
// when the package is imported. This allows us to "automatically" register the node
// with the registry at startup.
func init() {
	register.Register(engine.Node{
		ID:        ID,
		DependsOn: []string{node1.ID},
		Run:       run,
	})
}

// run the node's business logic and return a result that can be used
// by other nodes in the graph. It receives outputs from its dependencies (node1).
func run(deps map[string]engine.Result) (engine.Result, error) {
	// Extract the output from node1 using its type-safe helper
	n1, err := node1.FromDeps(deps)
	if err != nil {
		return engine.Result{}, err
	}

	fmt.Printf("  → Running %s (received: %q from node1)\n", ID, n1.Message)

	return engine.Result{
		ID: ID,
		Data: Output{
			Message: "node2c completed successfully",
		},
	}, nil
}
