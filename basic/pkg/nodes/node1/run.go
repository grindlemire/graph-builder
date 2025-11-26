package node1

import (
	"fmt"

	"github.com/grindlemire/graph-builder/basic/pkg/engine"
	"github.com/grindlemire/graph-builder/basic/pkg/register"
)

// ID is the unique identifier for the node. It is used to reference the node
// in the graph and to identify the node in the registry.
const ID = "node1"

// init registers the node with the registry. init is called automatically by Go
// when the package is imported. This allows us to "automatically" register the node
// with the registry at startup.
func init() {
	register.Register(engine.Node{
		ID: ID,
		// declare the dependencies for this node here
		// in this case, node1 has no dependencies
		DependsOn: []string{},
		Run:       run,
	})
}

// run the node's business logic and return a result that can be used
// by other nodes in the graph.
func run(deps map[string]engine.Result) (engine.Result, error) {
	fmt.Printf("  → Running %s (no dependencies)\n", ID)

	// business logic goes here to produce the Output
	output := Output{
		Message: "node1 completed successfully",
	}

	return engine.Result{
		ID:   ID,
		Data: output,
	}, nil
}
