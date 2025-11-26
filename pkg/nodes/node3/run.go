package node3

import (
	"fmt"

	"github.com/grindlemire/graph-builder/pkg/engine"
	"github.com/grindlemire/graph-builder/pkg/nodes/node2a"
	"github.com/grindlemire/graph-builder/pkg/nodes/node2b"
	"github.com/grindlemire/graph-builder/pkg/nodes/node2c"
	"github.com/grindlemire/graph-builder/pkg/register"
)

const ID = "node3"

func init() {
	register.Register(engine.Node{
		ID:        ID,
		DependsOn: []string{node2a.ID, node2b.ID, node2c.ID},
		Run:       run,
	})
}

func run(deps map[string]engine.Result) (engine.Result, error) {
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
