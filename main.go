package main

import (
	"flag"
	"log"

	"github.com/skeptic-ai/k8mpatible/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to kubeconfig. Leave empty for in-cluster configuration.")
	flag.Parse()

	var config *rest.Config
	client.CreateMergeGraph()
	graph, graphErr := client.LoadGraphFromYAML("merged_output.yaml")
	if graphErr != nil {
		log.Fatalf("Failed to load compatibility graph: %v", graphErr)
	}

	if *kubeconfig != "" {
		config, graphErr = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		config, graphErr = rest.InClusterConfig()
	}

	if graphErr != nil {
		log.Fatalf("Failed to create Kubernetes config: %v", graphErr)
	}
	if graphErr != nil {
		log.Fatalf("Failed to load compatibility graph: %v", graphErr)
	}
	clientset, clienterr := kubernetes.NewForConfig(config)
	if clienterr != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", clienterr)
	}
	client.ScanCluster(clientset, graph)

}
