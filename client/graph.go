package client

import (
	"github.com/Masterminds/semver/v3"
)

type Daemonset struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type KubernetesResource struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Type      string `yaml:"type"`
}

type Tool struct {
	Name                string               `yaml:"name"`
	KubernetesResources []KubernetesResource `yaml:"kubernetesResource"`
	DocUrl              string               `yaml:"docUrl"`
}

// Edge represents a relationship between two tools.
type Edge struct {
	SourceName              string              `yaml:"source"`
	SourceVersion           *semver.Version     `yaml:"sourceVersion"`
	DestinationName         string              `yaml:"destination"`
	DestinationVersionRange *semver.Constraints `yaml:"versionRange"`
	Compatible              bool                `yaml:"compatible"`
	Reason                  string              `yaml:"reason,omitempty"`
}

// Graph represents the entire compatibility graph.
type Graph struct {
	Nodes []*Tool `yaml:"nodes"`
	Edges []*Edge `yaml:"edges"`
}

const (
	Kubernetes = "Kubernetes"
)
