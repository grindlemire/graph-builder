package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/grindlemire/graph-builder/server/pkg/catalog"
	"github.com/grindlemire/graph-builder/server/pkg/engine"
	"github.com/grindlemire/graph-builder/server/pkg/nodes/node3"
	"github.com/grindlemire/graph-builder/server/pkg/nodes/node4"
)

func main() {
	// Create a engineBuilder from the node catalog (populated via init())
	engineBuilder := engine.NewBuilder(catalog.All())

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/graph/small", handleSmallGraph(engineBuilder))
	mux.HandleFunc("/graph/full", handleFullGraph(engineBuilder))
	mux.HandleFunc("/graph/custom", handleCustomGraph(engineBuilder))

	// Create server with explicit handler
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		fmt.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Run client tests
	runClientTests()

	// Shutdown server gracefully
	fmt.Println("\n" + "═══════════════════════════════════════")
	fmt.Println("All tests complete. Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	fmt.Println("Server stopped.")
}

func runClientTests() {
	client := &http.Client{Timeout: 10 * time.Second}

	endpoints := []struct {
		name string
		url  string
	}{
		{"Small Graph (node4 only)", "http://localhost:8080/graph/small"},
		{"Full Graph (node3 → all deps)", "http://localhost:8080/graph/full"},
		{"Custom Graph (node2a,node4)", "http://localhost:8080/graph/custom?nodes=node2a,node4"},
	}

	for _, ep := range endpoints {
		fmt.Println("\n" + "═══════════════════════════════════════")
		fmt.Printf("CLIENT: Requesting %s\n", ep.name)
		fmt.Printf("        URL: %s\n", ep.url)
		fmt.Println("═══════════════════════════════════════")

		resp, err := client.Get(ep.url)
		if err != nil {
			log.Printf("Request failed: %v", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("\nCLIENT: Response Status: %s\n", resp.Status)
		fmt.Printf("CLIENT: Response Body:\n%s\n", prettyJSON(body))
	}
}

func prettyJSON(data []byte) string {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return string(data)
	}
	pretty, err := json.MarshalIndent(obj, "  ", "  ")
	if err != nil {
		return string(data)
	}
	return "  " + string(pretty)
}

// handleSmallGraph runs a minimal graph: just node1 → node4
func handleSmallGraph(builder *engine.Builder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only request node4 - node1 is auto-resolved as a dependency
		e, err := builder.BuildFor(node4.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("\n=== /graph/small ===")
		e.PrettyPrint()

		if err := e.Run(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, e.Results())
	}
}

// handleFullGraph runs the full graph ending at node3 (which pulls in node2a, node2b, node2c, node1)
func handleFullGraph(builder *engine.Builder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only request node3 - all dependencies are auto-resolved
		e, err := builder.BuildFor(node3.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("\n=== /graph/full ===")
		e.PrettyPrint()

		if err := e.Run(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, e.Results())
	}
}

// handleCustomGraph builds a graph from query params: ?nodes=node2a,node4
func handleCustomGraph(builder *engine.Builder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodesParam := r.URL.Query().Get("nodes")
		if nodesParam == "" {
			http.Error(w, "missing 'nodes' query param (e.g. ?nodes=node2a,node4)", http.StatusBadRequest)
			return
		}

		// Parse comma-separated node IDs
		var targetNodes []string
		for _, n := range splitAndTrim(nodesParam) {
			if n != "" {
				targetNodes = append(targetNodes, n)
			}
		}

		e, err := builder.BuildFor(targetNodes...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("\n=== /graph/custom?nodes=%s ===\n", nodesParam)
		e.PrettyPrint()

		if err := e.Run(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, e.Results())
	}
}

func splitAndTrim(s string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			// Trim spaces
			for len(part) > 0 && part[0] == ' ' {
				part = part[1:]
			}
			for len(part) > 0 && part[len(part)-1] == ' ' {
				part = part[:len(part)-1]
			}
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	return result
}

func respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
