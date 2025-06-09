package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v2"
)

var test_graph = "./test-compatibility/test_graph.yaml"

func TestCheckCompatibility(t *testing.T) {
	one_9, _ := semver.NewVersion("1.9")
	one_9_range, _ := semver.NewConstraint(">=2.0, <3.0")
	one_10, _ := semver.NewVersion("1.10")
	one_10_range, _ := semver.NewConstraint(">=2.0, <2.26")
	one_11, _ := semver.NewVersion("1.11")
	one_11_range, _ := semver.NewConstraint(">3.0, <=3.3")
	two_25, _ := semver.NewVersion("2.25")
	three_0, _ := semver.NewVersion("3.0")
	three_3, _ := semver.NewVersion("3.3")
	// Sample graph for testing
	graph := &Graph{
		Edges: []*Edge{
			{
				SourceName:              "Istio",
				SourceVersion:           one_9,
				DestinationName:         "Prometheus",
				DestinationVersionRange: one_9_range,
				Compatible:              true,
			},
			{
				SourceName:              "Istio",
				SourceVersion:           one_10,
				DestinationName:         "Prometheus",
				DestinationVersionRange: one_10_range,
				Compatible:              false,
				Reason:                  "Known issue with X feature",
			},
			{
				SourceName:              "Istio",
				SourceVersion:           one_11,
				DestinationName:         "Prometheus",
				DestinationVersionRange: one_11_range,
				Compatible:              true,
			},
		},
	}

	tests := []struct {
		sourceName    string
		sourceVersion *semver.Version
		destName      string
		destVersion   *semver.Version
		expected      int
		reason        string
	}{
		{"Istio", one_9, "Prometheus", two_25, Compatible, ""},
		{"Istio", one_10, "Prometheus", two_25, NotCompatible, "Known issue with X feature"},
		{"Istio", one_9, "Prometheus", three_0, NotCompatible, "requires range >=2.0 <3.0"},
		{"Istio", one_11, "Prometheus", three_0, NotCompatible, "requires range >3.0 <=3.3"},
		{"Istio", one_11, "Prometheus", three_3, Compatible, ""},
	}

	for _, tt := range tests {
		compatible, reason := graph.CheckCompatibility(tt.sourceName, tt.sourceVersion, tt.destName, tt.destVersion)
		if compatible != tt.expected || reason != tt.reason {
			t.Errorf("CheckCompatibility(%s, %s, %s, %s) = %v, %s; want %v, %s",
				tt.sourceName, tt.sourceVersion, tt.destName, tt.destVersion, compatible, reason, tt.expected, tt.reason)
		}
	}
}

func TestYAMLUnmarshal(t *testing.T) {

	content, _ := os.ReadFile(test_graph)

	var graph Graph
	err := yaml.Unmarshal([]byte(content), &graph)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	expectedConstraint, _ := semver.NewConstraint(">=2.0, <3.0")

	if len(graph.Edges) != 2 || graph.Edges[0].DestinationVersionRange.String() != expectedConstraint.String() {
		fmt.Printf("%+v\n", graph)

		t.Errorf("Unexpected graph content: %+v", graph)
	}
}

func TestLoadGraphFromYAML(t *testing.T) {
	one_9, _ := semver.NewVersion("1.9")
	one_9_range, _ := semver.NewConstraint(">=2.0, <3.0")
	one_10, _ := semver.NewVersion("1.10")
	one_10_range, _ := semver.NewConstraint(">=2.0, <2.26")
	// one_11, _ := semver.NewVersion("1.11")
	// one_11_range, _ := semver.NewConstraint(">3.0, <=3.3")
	// two_25, _ := semver.NewVersion("2.25")
	// three_0, _ := semver.NewVersion("3.0")
	// three_3, _ := semver.NewVersion("3.3")
	// Create a temporary YAML file for testing

	// Load the graph from the YAML file
	graph, err := LoadGraphFromYAML(test_graph)
	if err != nil {
		t.Fatalf("Failed to load graph from YAML: %v", err)
	}

	// Expected graph structure
	expectedGraph := &Graph{
		Nodes: []*Tool{
			{Name: "Istio"},
			{Name: "Prometheus"},
			{Name: "Kubernetes", KubernetesResources: []KubernetesResource{{Name: "kube-proxy", Namespace: "kube-system", Type: "daemonset"}}},
		},
		Edges: []*Edge{
			{
				SourceName:              "Istio",
				SourceVersion:           one_9,
				DestinationName:         "Prometheus",
				DestinationVersionRange: one_9_range,
				Compatible:              true,
			},
			{
				SourceName:              "Istio",
				SourceVersion:           one_10,
				DestinationName:         "Prometheus",
				DestinationVersionRange: one_10_range,
				Compatible:              false,
				Reason:                  "Known issue with X feature",
			},
		},
	}

	if len(graph.Nodes) != len(expectedGraph.Nodes) || len(graph.Edges) != len(expectedGraph.Edges) {
		t.Errorf("Loaded graph does not match expected graph. Got nodes: %d, edges: %d; expected nodes: %d, edges: %d",
			len(graph.Nodes), len(graph.Edges), len(expectedGraph.Nodes), len(expectedGraph.Edges))
		return
	}

	for i, node := range graph.Nodes {
		if node.Name != expectedGraph.Nodes[i].Name {
			t.Errorf("Node mismatch at index %d. Got %s, expected %s", i, node.Name, expectedGraph.Nodes[i].Name)
		}
	}

	for i, edge := range graph.Edges {
		if edge.SourceName != expectedGraph.Edges[i].SourceName ||
			!edge.SourceVersion.Equal(expectedGraph.Edges[i].SourceVersion) ||
			edge.DestinationName != expectedGraph.Edges[i].DestinationName ||
			edge.DestinationVersionRange.String() != expectedGraph.Edges[i].DestinationVersionRange.String() ||
			edge.Compatible != expectedGraph.Edges[i].Compatible ||
			edge.Reason != expectedGraph.Edges[i].Reason {
			t.Errorf("Edge mismatch at index %d. Got %+v, expected %+v", i, edge, expectedGraph.Edges[i])
		}
	}
}
