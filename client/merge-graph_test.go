package client

import (
	"testing"
)

func TestCreateMergeGraph(t *testing.T) {
	graph, err := CreateMergeGraph()
	if err != nil {
		t.Fatalf("CreateMergeGraph() failed: %v", err)
	}

	if graph == nil {
		t.Fatal("CreateMergeGraph() returned nil graph")
	}

	// Basic validation that the graph has content
	if len(graph.Nodes) == 0 && len(graph.Edges) == 0 {
		t.Error("CreateMergeGraph() returned empty graph")
	}
}
