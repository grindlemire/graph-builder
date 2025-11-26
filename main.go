package main

import (
	"fmt"
	"log"

	"github.com/grindlemire/graph-builder/pkg/engine"
	"github.com/grindlemire/graph-builder/pkg/register"
)

func main() {
	// Build engine from registry (populated via init())
	e := engine.New(register.Registry())

	// Pretty print the graph structure
	e.PrettyPrint()

	// Execute in topological order
	if err := e.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("=== All nodes completed successfully ===")
}
