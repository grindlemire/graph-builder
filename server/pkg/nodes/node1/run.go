package node1

import (
	"fmt"

	"github.com/grindlemire/graph-builder/server/pkg/catalog"
	"github.com/grindlemire/graph-builder/server/pkg/engine"
)

// ID is the unique identifier for the node. It is used to reference the node
// in the graph and to identify the node in the catalog.
const ID = "node1"

// init registers the node with the catalog. init is called automatically by Go
// when the package is imported. This allows us to "automatically" register the node
// with the catalog at startup.
func init() {
	catalog.Register(engine.Node{
		ID:        ID,
		DependsOn: []string{},
		Run:       run,
	})
}

// run the node's business logic and return a result that can be used
// by other nodes in the graph.
func run(deps map[string]engine.Result) (engine.Result, error) {
	fmt.Printf("  → Running %s (no dependencies)\n", ID)

	// business logic goes here to produce the Output

	return engine.Result{
		ID: ID,
		Data: Output{
			Message: "node1 completed successfully",
		},
	}, nil
}
