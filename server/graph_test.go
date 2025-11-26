package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/grindlemire/graph-builder/server/pkg/catalog"
)

func TestGraphIntegrity(t *testing.T) {
	nodes := catalog.All()

	if len(nodes) == 0 {
		t.Fatal("no nodes registered in catalog")
	}

	t.Run("dependencies_exist", func(t *testing.T) {
		for id, node := range nodes {
			for _, dep := range node.DependsOn {
				if _, exists := nodes[dep]; !exists {
					t.Errorf("node %q declares dependency on %q which doesn't exist in catalog", id, dep)
				}
			}
		}
	})

	t.Run("no_cycles", func(t *testing.T) {
		visited := make(map[string]bool)
		recStack := make(map[string]bool)
		var cyclePath []string

		var hasCycle func(id string) bool
		hasCycle = func(id string) bool {
			visited[id] = true
			recStack[id] = true
			cyclePath = append(cyclePath, id)

			for _, dep := range nodes[id].DependsOn {
				if !visited[dep] {
					if hasCycle(dep) {
						return true
					}
				} else if recStack[dep] {
					cyclePath = append(cyclePath, dep)
					return true
				}
			}

			recStack[id] = false
			cyclePath = cyclePath[:len(cyclePath)-1]
			return false
		}

		for id := range nodes {
			if !visited[id] && hasCycle(id) {
				t.Errorf("cycle detected: %v", cyclePath)
				break
			}
		}
	})

	t.Run("fromdeps_matches_dependson", func(t *testing.T) {
		_, thisFile, _, _ := runtime.Caller(0)
		nodesDir := filepath.Join(filepath.Dir(thisFile), "pkg", "nodes")

		entries, err := os.ReadDir(nodesDir)
		if err != nil {
			t.Fatalf("failed to read nodes directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			nodeDir := filepath.Join(nodesDir, entry.Name())
			runFile := filepath.Join(nodeDir, "run.go")

			if _, err := os.Stat(runFile); os.IsNotExist(err) {
				continue
			}

			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, runFile, nil, 0)
			if err != nil {
				t.Errorf("failed to parse %s: %v", runFile, err)
				continue
			}

			analyzer := &nodeAnalyzer{declaredDeps: make(map[string]bool)}
			ast.Walk(analyzer, f)

			for _, used := range analyzer.usedDeps {
				if !analyzer.declaredDeps[used] {
					t.Errorf("%s/run.go: calls %s.FromDeps() but %s.ID is not in DependsOn",
						entry.Name(), used, used)
				}
			}
		}
	})
}

// nodeAnalyzer is a visitor that extracts dependency information from AST nodes.
type nodeAnalyzer struct {
	declaredDeps map[string]bool
	usedDeps     []string
}

func (a *nodeAnalyzer) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	a.checkDependsOn(n)
	a.checkFromDeps(n)
	return a
}

func (a *nodeAnalyzer) checkDependsOn(n ast.Node) {
	kv, ok := n.(*ast.KeyValueExpr)
	if !ok {
		return
	}
	key, ok := kv.Key.(*ast.Ident)
	if !ok || key.Name != "DependsOn" {
		return
	}
	arr, ok := kv.Value.(*ast.CompositeLit)
	if !ok {
		return
	}
	for _, elt := range arr.Elts {
		if sel, ok := elt.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok {
				a.declaredDeps[pkg.Name] = true
			}
		}
	}
}

func (a *nodeAnalyzer) checkFromDeps(n ast.Node) {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "FromDeps" {
		return
	}
	if pkg, ok := sel.X.(*ast.Ident); ok {
		a.usedDeps = append(a.usedDeps, pkg.Name)
	}
}
