# k8mpatible

Scan your Kubernetes cluster for tool version compatibility with your version of Kubernetes and with other tools, and provide the upgrade path.

## Overview

k8mpatible is a tool that helps you manage compatibility between different tools in your Kubernetes cluster. It scans your cluster to discover installed tools, checks their compatibility with each other and with your Kubernetes version, and provides guidance for upgrades.

## Running k8mpatible

### Prerequisites

- Go 1.23.4 or later
- Access to a Kubernetes cluster (either via kubeconfig or running in-cluster)

### Installation

```bash
# Clone the repository
git clone https://github.com/skeptic-ai/k8mpatible.git
cd k8mpatible

# Build the binary
go build -o k8mpatible .
```

### Usage

```bash
# Run with a specific kubeconfig file
./k8mpatible --kubeconfig=/path/to/your/kubeconfig

# Run in-cluster (when deployed as a pod in the cluster)
./k8mpatible
```

When run, k8mpatible will:

1. Scan your cluster to discover installed tools and their versions
2. Check compatibility between the discovered tools
3. Plan for potential Kubernetes upgrades and check compatibility with your tools
4. Output compatibility information and upgrade recommendations

## Adding New Tools

To add support for a new tool, you need to create a YAML file in the `client/compatibility` directory. The file should define:

1. The tool's name and how to discover it in the cluster
2. Compatibility relationships with other tools (especially Kubernetes)

### Tool Definition Format

Create a new YAML file in the `client/compatibility` directory with the following structure:

```yaml
nodes:
  - name: your-tool-name
    docUrl: "https://link-to-compatibility-docs"
    kubernetesResource:
      - name: resource-name
        namespace: resource-namespace
        type: deployment  # or daemonset
edges:
  - source: your-tool-name
    sourceVersion: 1.0  # Major.Minor version
    destination: Kubernetes  # or another tool
    versionRange: ">=1.24, <=1.27"  # Compatible Kubernetes versions
    compatible: true  # or false
    reason: "Optional explanation for compatibility status"
```

### Discovery Types

k8mpatible currently supports discovering tools via:
- Deployments
- DaemonSets

The tool extracts the version from the container image tag.

### Testing Your New Tool

After adding a new tool definition:

1. Run the tests:
   ```bash
   go test ./client/...
   ```

2. Run k8mpatible against a cluster with your tool installed to verify discovery and compatibility checking.

## Supported Tools

k8mpatible currently supports the following tools:

1. Kubernetes (core)
2. Argo CD
3. Argo Workflows
4. Cert Manager
5. External DNS
6. External Secrets
7. Flux
8. Istio
9. Linkerd

## Feature Ideas

- Add more tools
- Check compatibility of tools with node operating systems
- Support for more Kubernetes resource types for tool discovery

