package node1

import (
	"fmt"

	"github.com/grindlemire/graph-builder/pkg/engine"
)

type Output struct {
	Message string
}

func FromDeps(deps map[string]engine.Result) (Output, error) {
	result, ok := deps[ID]
	if !ok {
		return Output{}, fmt.Errorf("node1 result not found in deps")
	}

	output, ok := result.Data.(Output)
	if !ok {
		return Output{}, fmt.Errorf("invalid data type for node1")
	}

	return output, nil
}

