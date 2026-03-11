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

// LabelVersionDiscoverer finds resources by label selector across all namespaces.
type LabelVersionDiscoverer struct {
	Labels       map[string]string
	ResourceType string
	Image        string
}

func (d *LabelVersionDiscoverer) DiscoverVersion(clientset *kubernetes.Clientset) (*semver.Version, error) {
	selector := buildLabelSelector(d.Labels)

	switch d.ResourceType {
	case "deployment":
		deployments, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return nil, err
		}
		if len(deployments.Items) == 0 {
			return nil, fmt.Errorf("no deployments found with labels %s", selector)
		}
		return getImageFromContainers(deployments.Items[0].Spec.Template.Spec.Containers, d.Image)

	case "daemonset":
		daemonsets, err := clientset.AppsV1().DaemonSets("").List(context.Background(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return nil, err
		}
		if len(daemonsets.Items) == 0 {
			return nil, fmt.Errorf("no daemonsets found with labels %s", selector)
		}
		return getImageFromContainers(daemonsets.Items[0].Spec.Template.Spec.Containers, d.Image)
	}

	return nil, fmt.Errorf("unsupported resource type: %s", d.ResourceType)
}

func buildLabelSelector(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

// ImageVersionDiscoverer finds resources by matching container image path (registry-agnostic).
type ImageVersionDiscoverer struct {
	Image        string
	ResourceType string
}

func (d *ImageVersionDiscoverer) DiscoverVersion(clientset *kubernetes.Clientset) (*semver.Version, error) {
	switch d.ResourceType {
	case "deployment":
		deployments, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, deployment := range deployments.Items {
			if version, err := getVersionFromContainersByImage(deployment.Spec.Template.Spec.Containers, d.Image); err == nil {
				return version, nil
			}
		}
		return nil, fmt.Errorf("no deployment found with image containing %s", d.Image)

	case "daemonset":
		daemonsets, err := clientset.AppsV1().DaemonSets("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, daemonset := range daemonsets.Items {
			if version, err := getVersionFromContainersByImage(daemonset.Spec.Template.Spec.Containers, d.Image); err == nil {
				return version, nil
			}
		}
		return nil, fmt.Errorf("no daemonset found with image containing %s", d.Image)
	}

	return nil, fmt.Errorf("unsupported resource type: %s", d.ResourceType)
}

// getVersionFromContainersByImage checks if any container's image contains the given image path
// and extracts the version from its tag.
func getVersionFromContainersByImage(containers []corev1.Container, imagePath string) (*semver.Version, error) {
	for _, container := range containers {
		// Split off the tag to compare just the image path
		imageRef := container.Image
		tagSep := strings.LastIndex(imageRef, ":")
		if tagSep == -1 {
			continue
		}
		imageWithoutTag := imageRef[:tagSep]
		if strings.Contains(imageWithoutTag, imagePath) {
			tag := imageRef[tagSep+1:]
			return semver.NewVersion(tag)
		}
	}
	return nil, fmt.Errorf("no container with image matching %s", imagePath)
}

type DeploymentVersionDiscoverer struct {
	Name      string
	Namespace string
	Image     string
}

func (d *DeploymentVersionDiscoverer) DiscoverVersion(clientset *kubernetes.Clientset) (*semver.Version, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces.Items {
		deployments, err := clientset.AppsV1().Deployments(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, deployment := range deployments.Items {
			if strings.Contains(deployment.Name, d.Name) {
				return getImageFromContainers(deployment.Spec.Template.Spec.Containers, d.Image)
			}
		}
	}
	return nil, fmt.Errorf("deployment %s not found", d.Name)
}

type DaemonSetVersionDiscoverer struct {
	Name      string
	Namespace string
	Image     string
}

func (d *DaemonSetVersionDiscoverer) DiscoverVersion(clientset *kubernetes.Clientset) (*semver.Version, error) {
	daemonset, err := clientset.AppsV1().DaemonSets(d.Namespace).Get(context.Background(), d.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return getImageFromContainers(daemonset.Spec.Template.Spec.Containers, d.Image)
}

// FallbackVersionDiscoverer tries labels, then image, then name-based discovery.
type FallbackVersionDiscoverer struct {
	Resource KubernetesResource
}

func (f *FallbackVersionDiscoverer) DiscoverVersion(clientset *kubernetes.Clientset) (*semver.Version, error) {
	// Try label-based discovery first
	if len(f.Resource.Labels) > 0 {
		discoverer := &LabelVersionDiscoverer{
			Labels:       f.Resource.Labels,
			ResourceType: f.Resource.Type,
			Image:        f.Resource.Image,
		}
		if version, err := discoverer.DiscoverVersion(clientset); err == nil {
			return version, nil
		}
	}

	// Try image-based discovery second
	if f.Resource.Image != "" {
		discoverer := &ImageVersionDiscoverer{
			Image:        f.Resource.Image,
			ResourceType: f.Resource.Type,
		}
		if version, err := discoverer.DiscoverVersion(clientset); err == nil {
			return version, nil
		}
	}

	// Fall back to name-based discovery
	switch f.Resource.Type {
	case "deployment":
		return (&DeploymentVersionDiscoverer{Name: f.Resource.Name, Namespace: f.Resource.Namespace, Image: f.Resource.Image}).DiscoverVersion(clientset)
	case "daemonset":
		return (&DaemonSetVersionDiscoverer{Name: f.Resource.Name, Namespace: f.Resource.Namespace, Image: f.Resource.Image}).DiscoverVersion(clientset)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", f.Resource.Type)
	}
}

func NewVersionDiscoverer(resource KubernetesResource) VersionDiscoverer {
	return &FallbackVersionDiscoverer{Resource: resource}
}

func getImageFromContainers(containers []corev1.Container, imageHint ...string) (*semver.Version, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("no containers found")
	}

	var image string
	if len(imageHint) > 0 && imageHint[0] != "" {
		// Match the container whose image path contains the hint
		for _, c := range containers {
			imageRef := c.Image
			tagSep := strings.LastIndex(imageRef, ":")
			imageWithoutTag := imageRef
			if tagSep != -1 {
				imageWithoutTag = imageRef[:tagSep]
			}
			if strings.Contains(imageWithoutTag, imageHint[0]) {
				image = c.Image
				break
			}
		}
		if image == "" {
			// Hint didn't match (e.g. ECR mirror changed the image path); fall back to first container
			image = containers[0].Image
		}
	} else {
		image = containers[0].Image
	}

	imageParts := strings.Split(image, ":")
	if len(imageParts) < 2 {
		return nil, fmt.Errorf("unable to find image tag")
	}
	version, err := semver.NewVersion(imageParts[len(imageParts)-1])
	if err != nil {
		return nil, err
	}
	return version, nil
}

type Incompatibility struct {
	Message  string `json:"message"`
	ToolName string `json:"tool_name"`
}

type DiscoveredTool struct {
	Name                   string            `json:"name"`
	Version                *semver.Version   `json:"version"`
	DocUrl                 string            `json:"docUrl" yaml:"docUrl"`
	UpgradeIncompatibility []Incompatibility `json:"upgrade_incompatibility"`
	CurrentIncompatibility []Incompatibility `json:"current_incompatibility"`
}
