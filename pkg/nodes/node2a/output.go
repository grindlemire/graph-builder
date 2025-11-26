package node2a

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
		return Output{}, fmt.Errorf("node2a result not found in deps")
	}

	output, ok := result.Data.(Output)
	if !ok {
		return Output{}, fmt.Errorf("invalid data type for node2a")
	}

	return output, nil
}

