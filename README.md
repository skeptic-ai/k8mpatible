# k8mpatible

Scan your Kubernetes cluster for tool version compatibility with your version of Kubernetes and with other tools, and provide the upgrade path.

## Overview

k8mpatible is a tool that helps you manage compatibility between different tools in your Kubernetes cluster. It scans your cluster to discover installed tools, checks their compatibility with each other and with your Kubernetes version, and provides guidance for upgrades.

## Running k8mpatible

### Prerequisites

- Go 1.23.4 or later (only if building from source)
- Access to a Kubernetes cluster (either via kubeconfig or running in-cluster)

### Installation

#### Option 1: Download pre-built binary (recommended)

Download the latest release from the [GitHub Releases page](https://github.com/skeptic-ai/k8mpatible/releases/latest) for your platform.

Or use these commands to download and install the latest release:

```bash
# Set the version (or get the latest version automatically)
VERSION=$(curl -s https://api.github.com/repos/skeptic-ai/k8mpatible/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

# For Linux (x86_64)
curl -L https://github.com/skeptic-ai/k8mpatible/releases/latest/download/k8mpatible_${VERSION#v}_Linux_x86_64.tar.gz | tar xz
sudo mv k8mpatible /usr/local/bin/

# For macOS (x86_64)
curl -L https://github.com/skeptic-ai/k8mpatible/releases/latest/download/k8mpatible_${VERSION#v}_Darwin_x86_64.tar.gz | tar xz
sudo mv k8mpatible /usr/local/bin/

# For macOS (arm64/Apple Silicon)
curl -L https://github.com/skeptic-ai/k8mpatible/releases/latest/download/k8mpatible_${VERSION#v}_Darwin_arm64.tar.gz | tar xz
sudo mv k8mpatible /usr/local/bin/

# For Windows (PowerShell) - Run in two steps
$VERSION = (Invoke-RestMethod -Uri "https://api.github.com/repos/skeptic-ai/k8mpatible/releases/latest").tag_name.TrimStart("v")
Invoke-WebRequest -Uri "https://github.com/skeptic-ai/k8mpatible/releases/latest/download/k8mpatible_${VERSION}_Windows_x86_64.zip" -OutFile k8mpatible.zip
Expand-Archive -Path k8mpatible.zip -DestinationPath .
# Move k8mpatible.exe to a directory in your PATH
```

You can also manually download the appropriate binary for your platform from the [GitHub Releases page](https://github.com/skeptic-ai/k8mpatible/releases/latest).

#### Option 2: Build from source

```bash
# Clone the repository
git clone https://github.com/skeptic-ai/k8mpatible.git
cd k8mpatible

# Build the binary
go build -o k8mpatible .
```

### Usage

```bash
# Run with default kubeconfig location ($HOME/.kube/config)
./k8mpatible

# Run with a specific kubeconfig file
./k8mpatible --kubeconfig=/path/to/your/kubeconfig

# Export scan results to a YAML file
./k8mpatible --output=scan-results.yaml

# Specify both kubeconfig and output file
./k8mpatible --kubeconfig=/path/to/your/kubeconfig --output=scan-results.yaml
```

When run, k8mpatible will:

1. Scan your cluster to discover installed tools and their versions
2. Check compatibility between the discovered tools
3. Plan for potential Kubernetes upgrades and check compatibility with your tools
4. Output compatibility information and upgrade recommendations
5. Optionally export the scan results to a YAML file if the `--output` flag is specified

### YAML Export Format

When using the `--output` flag, k8mpatible will export the scan results to a YAML file with the following structure:

```yaml
tools:
  - name: tool-name
    version: x.y.z
    docUrl: https://link-to-tool-compatibility-documentation
    current_incompatibility:
      - message: "Reason for incompatibility"
        tool_name: "Incompatible tool name"
    upgrade_incompatibility:
      - message: "Reason for upgrade incompatibility"
        tool_name: "Incompatible tool name after upgrade"
summary:
  totalTools: 5
  incompatibleTools: 1
  upgradeIncompatibleTools: 2
```

The YAML file includes:
- A list of all discovered tools with their versions
- Any current incompatibilities between tools
- Any incompatibilities that would occur after a Kubernetes upgrade
- A summary section with statistics about the scan results

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
10. Aws LB controller
11. Stakater Reloader
12. datadog agent

## Feature Ideas

- Add more tools
- Check compatibility of tools with node operating systems
- Support for more Kubernetes resource types for tool discovery
