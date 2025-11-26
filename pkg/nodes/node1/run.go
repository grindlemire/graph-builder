package node1

import (
	"fmt"

	"github.com/grindlemire/graph-builder/pkg/engine"
	"github.com/grindlemire/graph-builder/pkg/register"
)

const ID = "node1"

func init() {
	register.Register(engine.Node{
		ID:        ID,
		DependsOn: []string{},
		Run:       run,
	})
}

func run(deps map[string]engine.Result) (engine.Result, error) {
	fmt.Printf("  → Running %s (no dependencies)\n", ID)

	return engine.Result{
		ID: ID,
		Data: Output{
			Message: "node1 completed successfully",
		},
	}, nil
}

