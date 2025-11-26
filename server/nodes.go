package main

import (
	// Each node package is owned by a different team.
	// Adding a new node = create a new package + add one import line here.
	// Nodes self-register to the catalog via init()
	_ "github.com/grindlemire/graph-builder/server/pkg/nodes/node1"
	_ "github.com/grindlemire/graph-builder/server/pkg/nodes/node2a"
	_ "github.com/grindlemire/graph-builder/server/pkg/nodes/node2b"
	_ "github.com/grindlemire/graph-builder/server/pkg/nodes/node2c"
	_ "github.com/grindlemire/graph-builder/server/pkg/nodes/node3"
	_ "github.com/grindlemire/graph-builder/server/pkg/nodes/node4"
)
