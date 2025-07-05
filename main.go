package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/skeptic-ai/k8mpatible/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to kubeconfig. Leave empty to use the default location ($HOME/.kube/config).")
	outputFile := flag.String("output", "", "Path to output YAML file for scan results. If not specified, results will only be logged.")
	flag.Parse()

	var config *rest.Config
	graph, graphErr := client.CreateMergeGraph()
	if graphErr != nil {
		log.Fatalf("Failed to create compatibility graph: %v", graphErr)
	}

	// Try to build config from specified kubeconfig path, default path, or in-cluster config
	if *kubeconfig != "" {
		// Use specified kubeconfig path
		config, graphErr = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		// Try to use default kubeconfig path
		kubeconfigPath := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, graphErr = clientcmd.BuildConfigFromFlags("", kubeconfigPath)

		// If that fails, try in-cluster config
		if graphErr != nil {
			log.Printf("Could not load kubeconfig from default location: %v", graphErr)
			log.Printf("Trying in-cluster configuration...")
			config, graphErr = rest.InClusterConfig()
		}
	}

	if graphErr != nil {
		log.Fatalf("Failed to create Kubernetes config: %v", graphErr)
	}
	clientset, clienterr := kubernetes.NewForConfig(config)
	if clienterr != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", clienterr)
	}

	// Run the scan
	scanResults := client.ScanCluster(clientset, graph)

	// Export scan results to YAML and print to stdout
	yamlData, err := client.ExportScanResultsToYAML(scanResults, *outputFile)
	if err != nil {
		log.Fatalf("Failed to export scan results to YAML: %v", err)
	}

	// Print the YAML data to stdout
	fmt.Println(string(yamlData))

	// Log if the results were also exported to a file
	if *outputFile != "" {
		log.Printf("Scan results exported to %s", *outputFile)
	}
}
