package client

import (
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/kubernetes"
)

func ScanCluster(clientset *kubernetes.Clientset, graph *Graph) []DiscoveredTool {

	discoveredTools := discoverTools(clientset, graph)

	checkDiscoveredToolsCompatiblity(discoveredTools, graph)

	// Plan kubernetes upgrade
	k8sVersion := GetActiveVersion(discoveredTools, Kubernetes)
	nextVersion := k8sVersion.IncMinor()
	planToolVersion(Kubernetes, &nextVersion, graph, discoveredTools)
	Logger.Info("tool upgrade plan completed", "tools", discoveredTools)
	return discoveredTools
}

// ScanResults represents the structure of the scan results
type ScanResults struct {
	Tools   []DiscoveredTool `yaml:"tools"`
	Summary struct {
		TotalTools               int `yaml:"totalTools"`
		IncompatibleTools        int `yaml:"incompatibleTools"`
		UpgradeIncompatibleTools int `yaml:"upgradeIncompatibleTools"`
	} `yaml:"summary"`
}

// FormatScanResultsAsYAML formats the scan results as YAML and returns the YAML data
func FormatScanResultsAsYAML(tools []DiscoveredTool) ([]byte, *ScanResults, error) {
	// Create the scan results
	results := &ScanResults{
		Tools: tools,
	}

	// Calculate summary statistics
	results.Summary.TotalTools = len(tools)

	incompatibleTools := 0
	upgradeIncompatibleTools := 0

	for _, tool := range tools {
		if len(tool.CurrentIncompatibility) > 0 {
			incompatibleTools++
		}
		if len(tool.UpgradeIncompatibility) > 0 {
			upgradeIncompatibleTools++
		}
	}

	results.Summary.IncompatibleTools = incompatibleTools
	results.Summary.UpgradeIncompatibleTools = upgradeIncompatibleTools

	// Marshal the scan results to YAML
	yamlData, err := yaml.Marshal(results)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal scan results to YAML: %v", err)
	}

	return yamlData, results, nil
}

// ExportScanResultsToYAML exports the scan results to a YAML file and returns the YAML data
func ExportScanResultsToYAML(tools []DiscoveredTool, outputPath string) ([]byte, error) {
	// Format the scan results as YAML
	yamlData, _, err := FormatScanResultsAsYAML(tools)
	if err != nil {
		return nil, err
	}

	// If an output path is provided, write the YAML data to the output file
	if outputPath != "" {
		err = os.WriteFile(outputPath, yamlData, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write scan results to file: %v", err)
		}
		Logger.Info("scan results exported to YAML", "path", outputPath)
	}

	return yamlData, nil
}

func discoverTools(clientset *kubernetes.Clientset, graph *Graph) []DiscoveredTool {
	var discoveredTools []DiscoveredTool

	for _, node := range graph.Nodes {
		for _, resource := range node.KubernetesResources {
			discoverer := NewVersionDiscoverer(resource.Type, resource.Name, resource.Namespace)
			version, err := discoverer.DiscoverVersion(clientset)
			if err != nil {
				continue
			}
			discoveredTools = append(discoveredTools, DiscoveredTool{
				Name:                   node.Name,
				Version:                version,
				DocUrl:                 node.DocUrl,
				UpgradeIncompatibility: []Incompatibility{},
				CurrentIncompatibility: []Incompatibility{},
			})
		}
	}

	return discoveredTools
}

func checkDiscoveredToolsCompatiblity(discoveredTools []DiscoveredTool, graph *Graph) {

	for i := 0; i < len(discoveredTools); i++ {
		for j := i + 1; j < len(discoveredTools); j++ {
			toolA := discoveredTools[i]
			toolB := discoveredTools[j]
			compatibility, reason := graph.CheckCompatibility(toolA.Name, toolA.Version, toolB.Name, toolB.Version)
			if compatibility == NotCompatible {
				message := fmt.Sprintf("%s (version %s) is not compatible with %s (version %s). Reason: %s\n", toolA.Name, toolA.Version, toolB.Name, toolB.Version, reason)
				Logger.Info(message)
				incompatiblity := Incompatibility{reason, toolB.Name}
				toolA.CurrentIncompatibility = append(toolA.CurrentIncompatibility, incompatiblity)
				discoveredTools[i].CurrentIncompatibility = toolA.CurrentIncompatibility

			} else if compatibility == Compatible {
				message := fmt.Sprintf("%s (version %s) is compatible with %s (version %s)\n", toolA.Name, toolA.Version, toolB.Name, toolB.Version)
				Logger.Info(message)
			} else if compatibility == Unknown {
				message := fmt.Sprintf("Compatibility between %s (version %s) and %s (version %s) is unknown\n", toolA.Name, toolA.Version, toolB.Name, toolB.Version)
				Logger.Info(message)
			}
		}
	}
}

func planToolVersion(toolName string, toolVersion *semver.Version, graph *Graph, discoveredTools []DiscoveredTool) {

	for idx, tool := range discoveredTools {
		if tool.Name == toolName {
			continue
		}
		compatibility, reason := graph.CheckCompatibility(tool.Name, tool.Version, Kubernetes, toolVersion)
		if compatibility == NotCompatible {
			message := fmt.Sprintf("%s (version %s) is not compatible with %s (version %s). Reason: %s\n", toolName, toolVersion, tool.Name, tool.Version, reason)
			Logger.Info(message)
			incompatiblity := Incompatibility{reason, tool.Name}
			tool.UpgradeIncompatibility = append(tool.UpgradeIncompatibility, incompatiblity)
			discoveredTools[idx].UpgradeIncompatibility = tool.UpgradeIncompatibility
		} else if compatibility == Compatible {
			message := fmt.Sprintf("%s (version %s) is compatible with %s (version %s)\n", Kubernetes, toolVersion, tool.Name, tool.Version)
			Logger.Info(message)
		} else if compatibility == Unknown {
			message := fmt.Sprintf("Compatibility between %s (version %s) and %s (version %s) is unknown\n", Kubernetes, toolVersion, tool.Name, tool.Version)
			Logger.Info(message)
		}
	}
}
