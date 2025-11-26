package register

import "github.com/grindlemire/graph-builder/pkg/engine"

var registry = make(map[string]engine.Node)

// Register adds a node to the global registry.
// Called from init() functions in check packages.
func Register(node engine.Node) {
	if _, exists := registry[node.ID]; exists {
		// panic here because this is called in an init function and we want to fail fast
		panic("duplicate node registration: " + node.ID)
	}
	registry[node.ID] = node
}

// Registry returns all registered nodes
func Registry() map[string]engine.Node {
	return registry
}
