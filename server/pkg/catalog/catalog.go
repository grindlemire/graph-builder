package catalog

import "github.com/grindlemire/graph-builder/server/pkg/engine"

// Global catalog of all available nodes
var nodes = make(map[string]engine.Node)

// Register adds a node to the catalog.
// Called from init() functions in node packages.
func Register(node engine.Node) {
	if _, exists := nodes[node.ID]; exists {
		panic("duplicate node registration: " + node.ID)
	}
	nodes[node.ID] = node
}

// Get returns a node by ID
func Get(id string) (engine.Node, bool) {
	n, ok := nodes[id]
	return n, ok
}

// All returns the complete node catalog
func All() map[string]engine.Node {
	return nodes
}

