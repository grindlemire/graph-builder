package node4

import (
	"fmt"

	"github.com/grindlemire/graph-builder/pkg/engine"
	"github.com/grindlemire/graph-builder/pkg/nodes/node1"
	"github.com/grindlemire/graph-builder/pkg/register"
)

const ID = "node4"

func init() {
	register.Register(engine.Node{
		ID:        ID,
		DependsOn: []string{node1.ID},
		Run:       run,
	})
}

func run(deps map[string]engine.Result) (engine.Result, error) {
	n1, err := node1.FromDeps(deps)
	if err != nil {
		return engine.Result{}, err
	}

	fmt.Printf("  → Running %s (received: %q from node1)\n", ID, n1.Message)

	return engine.Result{
		ID: ID,
		Data: Output{
			Message: "node4 completed successfully",
		},
	}, nil
}
