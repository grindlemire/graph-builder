package engine

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Result holds the output of a node execution
type Result struct {
	ID   string
	Data any
}

// RunFunc is the signature for a node's execution function.
// It receives results from all dependencies.
type RunFunc func(deps map[string]Result) (Result, error)

// Node represents a single node in the dependency graph
type Node struct {
	ID        string
	DependsOn []string
	Run       RunFunc
}

// Engine manages the dependency graph and execution
type Engine struct {
	nodes   map[string]Node
	results map[string]Result
	mu      sync.RWMutex
}

// New creates an engine from a registry of nodes
func New(registry map[string]Node) *Engine {
	return &Engine{
		nodes:   registry,
		results: make(map[string]Result),
	}
}

// PrettyPrint outputs a visual representation of the dependency graph
func (e *Engine) PrettyPrint() {
	fmt.Println("┌─────────────────────────────────────┐")
	fmt.Println("│         Dependency Graph            │")
	fmt.Println("└─────────────────────────────────────┘")

	// Get sorted node IDs for consistent output
	ids := make([]string, 0, len(e.nodes))
	for id := range e.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Build reverse map (who depends on me)
	dependents := make(map[string][]string)
	for _, node := range e.nodes {
		for _, dep := range node.DependsOn {
			dependents[dep] = append(dependents[dep], node.ID)
		}
	}

	for _, id := range ids {
		node := e.nodes[id]
		fmt.Printf("\n  ◉ %s\n", id)

		if len(node.DependsOn) > 0 {
			sort.Strings(node.DependsOn)
			fmt.Printf("    ├─ depends on: %s\n", strings.Join(node.DependsOn, ", "))
		} else {
			fmt.Printf("    ├─ depends on: (none - root node)\n")
		}

		if deps, ok := dependents[id]; ok && len(deps) > 0 {
			sort.Strings(deps)
			fmt.Printf("    └─ required by: %s\n", strings.Join(deps, ", "))
		} else {
			fmt.Printf("    └─ required by: (none - leaf node)\n")
		}
	}

	// Show execution levels
	levels, err := e.topoSortLevels()
	if err != nil {
		fmt.Printf("\n  ⚠ Error computing levels: %v\n", err)
		return
	}

	fmt.Printf("\n\n")
	fmt.Println("┌─────────────────────────────────────┐")
	fmt.Println("│         Execution Levels            │")
	fmt.Println("└─────────────────────────────────────┘")

	for i, level := range levels {
		sort.Strings(level)
		parallel := ""
		if len(level) > 1 {
			parallel = " (parallel)"
		}
		fmt.Printf("\n  Level %d%s:\n", i, parallel)
		for _, id := range level {
			fmt.Printf("    → %s\n", id)
		}
	}
	fmt.Println()
}

// Run executes all nodes in parallel where possible.
// Nodes are grouped into levels based on dependencies.
// All nodes in a level run concurrently, levels execute sequentially.
func (e *Engine) Run() error {
	levels, err := e.topoSortLevels()
	if err != nil {
		return err
	}

	fmt.Printf("\n\n")
	fmt.Println("┌─────────────────────────────────────┐")
	fmt.Println("│           Executing Graph           │")
	fmt.Println("└─────────────────────────────────────┘")

	for levelNum, level := range levels {
		sort.Strings(level)
		if len(level) > 1 {
			fmt.Printf("\n⚡ Level %d: executing %d nodes in parallel [%s]\n", levelNum, len(level), strings.Join(level, ", "))
		} else {
			fmt.Printf("\n◆ Level %d: executing [%s]\n", levelNum, level[0])
		}

		var wg sync.WaitGroup
		errCh := make(chan error, len(level))

		for _, id := range level {
			wg.Add(1)
			go func(nodeID string) {
				defer wg.Done()

				node := e.nodes[nodeID]

				// Gather dependency results (safe to read, deps already complete)
				depResults := make(map[string]Result)
				e.mu.RLock()
				for _, depID := range node.DependsOn {
					// this is storing values so we don't need to lock
					// the result from the map
					depResults[depID] = e.results[depID]
				}
				e.mu.RUnlock()

				// Execute node
				result, err := node.Run(depResults)
				if err != nil {
					errCh <- fmt.Errorf("node %s failed: %w", nodeID, err)
					return
				}

				e.mu.Lock()
				e.results[nodeID] = result
				e.mu.Unlock()

				fmt.Printf("  ✓ %s completed\n", nodeID)
			}(id)
		}

		wg.Wait()
		close(errCh)

		// Return first error encountered
		if err := <-errCh; err != nil {
			return err
		}
	}

	return nil
}

// Results returns all collected results after execution
func (e *Engine) Results() map[string]Result {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.results
}

// Builder constructs engines from a node catalog with automatic dependency resolution
type Builder struct {
	catalog map[string]Node
}

// NewBuilder creates a builder from a node catalog
func NewBuilder(catalog map[string]Node) *Builder {
	return &Builder{catalog: catalog}
}

// BuildFor creates an engine with the specified target nodes and ALL their transitive dependencies.
// Just specify the terminal nodes you need - dependencies are resolved automatically.
func (b *Builder) BuildFor(targetNodeIDs ...string) (*Engine, error) {
	needed := make(map[string]Node)

	var resolve func(id string) error
	resolve = func(id string) error {
		if _, already := needed[id]; already {
			return nil
		}
		node, ok := b.catalog[id]
		if !ok {
			return fmt.Errorf("unknown node: %s", id)
		}
		needed[id] = node
		for _, dep := range node.DependsOn {
			if err := resolve(dep); err != nil {
				return err
			}
		}
		return nil
	}

	for _, id := range targetNodeIDs {
		if err := resolve(id); err != nil {
			return nil, err
		}
	}

	return New(needed), nil
}

// topoSortLevels returns nodes grouped into levels.
// Nodes in the same level have no dependencies on each other and can run in parallel.
func (e *Engine) topoSortLevels() ([][]string, error) {
	// Build in-degree map
	inDegree := make(map[string]int)
	for id := range e.nodes {
		inDegree[id] = 0
	}
	for _, node := range e.nodes {
		for _, dep := range node.DependsOn {
			if _, exists := e.nodes[dep]; !exists {
				return nil, fmt.Errorf("node %s depends on unknown node %s", node.ID, dep)
			}
		}
		inDegree[node.ID] = len(node.DependsOn)
	}

	// Find nodes with no dependencies (first level)
	var currentLevel []string
	for id, degree := range inDegree {
		if degree == 0 {
			currentLevel = append(currentLevel, id)
		}
	}

	// Build reverse adjacency (who depends on me)
	dependents := make(map[string][]string)
	for _, node := range e.nodes {
		for _, dep := range node.DependsOn {
			dependents[dep] = append(dependents[dep], node.ID)
		}
	}

	// Process level by level
	var levels [][]string
	processed := 0

	for len(currentLevel) > 0 {
		levels = append(levels, currentLevel)
		processed += len(currentLevel)

		var nextLevel []string
		for _, id := range currentLevel {
			for _, dependent := range dependents[id] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					nextLevel = append(nextLevel, dependent)
				}
			}
		}
		currentLevel = nextLevel
	}

	if processed != len(e.nodes) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return levels, nil
}
