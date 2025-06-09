package client

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
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

func discoverTools(clientset *kubernetes.Clientset, graph *Graph) []DiscoveredTool {
	var discoveredTools []DiscoveredTool

	for _, node := range graph.Nodes {
		for _, resource := range node.KubernetesResources {
			discoverer := NewVersionDiscoverer(resource.Type, resource.Name, resource.Namespace)
			version, err := discoverer.DiscoverVersion(clientset)
			if err != nil {
				continue
			}
			discoveredTools = append(discoveredTools, DiscoveredTool{node.Name, version, []Incompatibility{}, []Incompatibility{}})
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
