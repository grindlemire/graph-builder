package node3

import (
	"fmt"

	"github.com/grindlemire/graph-builder/server/pkg/engine"
)

// Output is the output of the node that other nodes in the graph can use.
type Output struct {
	Message string
}

// FromDeps is a helper function that returns the Output for this node
// from the set of dependencies. This is used by other nodes to easily
// parse this node's output.
func FromDeps(deps map[string]engine.Result) (Output, error) {
	result, ok := deps[ID]
	if !ok {
		return Output{}, fmt.Errorf("node3 result not found in deps")
	}

	output, ok := result.Data.(Output)
	if !ok {
		return Output{}, fmt.Errorf("invalid data type for node3")
	}

	return output, nil
}
