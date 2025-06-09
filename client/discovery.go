package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type VersionDiscoverer interface {
	DiscoverVersion(*kubernetes.Clientset) (*semver.Version, error)
}

type DeploymentVersionDiscoverer struct {
	Name      string
	Namespace string
}

func (d *DeploymentVersionDiscoverer) DiscoverVersion(clientset *kubernetes.Clientset) (*semver.Version, error) {
	// Get all namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error getting namespaces %s", err)
		return nil, err
	}
	for _, namespace := range namespaces.Items {
		// list deployments in the namespace
		deployments, err := clientset.AppsV1().Deployments(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Println("error getting deployments")
			return nil, err
		}
		for _, deployment := range deployments.Items {
			// if deployment name contains the name of the deployment we are looking for
			if strings.Contains(deployment.Name, d.Name) {
				image, err := getImageFromContainers(deployment.Spec.Template.Spec.Containers)
				if err != nil {
					fmt.Println("error getting image from containers")
					return nil, err
				}
				return image, nil
			}
		}
	}
	return nil, fmt.Errorf("deployment not found")
}

type DaemonSetVersionDiscoverer struct {
	Name      string
	Namespace string
}

func (d *DaemonSetVersionDiscoverer) DiscoverVersion(clientset *kubernetes.Clientset) (*semver.Version, error) {
	daemonset, err := clientset.AppsV1().DaemonSets(d.Namespace).Get(context.Background(), d.Name, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return nil, err
	}
	image, err := getImageFromContainers(daemonset.Spec.Template.Spec.Containers)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func getImageFromContainers(containers []corev1.Container) (*semver.Version, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("no containers found")
	}
	image := containers[0].Image
	imageParts := strings.Split(image, ":")
	if len(imageParts) < 2 {
		return nil, fmt.Errorf("unable to find image tag")
	}
	version, err := semver.NewVersion(imageParts[len(imageParts)-1])
	if err != nil {
		fmt.Printf("error parsing version: %v\n", err)
		return nil, err
	}
	return version, nil
}

func NewVersionDiscoverer(resourceType, name, namespace string) VersionDiscoverer {
	switch resourceType {
	case "deployment":
		return &DeploymentVersionDiscoverer{Name: name, Namespace: namespace}
	case "daemonset":
		return &DaemonSetVersionDiscoverer{Name: name, Namespace: namespace}
	default:
		return nil // or return an error indicating unsupported resource type
	}
}

type Incompatibility struct {
	Message  string `json:"message"`
	ToolName string `json:"tool_name"`
}

type DiscoveredTool struct {
	Name                   string            `json:"name"`
	Version                *semver.Version   `json:"version"`
	UpgradeIncompatibility []Incompatibility `json:"upgrade_incompatibility"`
	CurrentIncompatibility []Incompatibility `json:"current_incompatibility"`
}
