#!/usr/bin/env bash
# E2E test suite for k8mpatible.
# Expects a KIND cluster named "k8mpatible-e2e" to already be running
# (created by helm/kind-action in CI, or manually for local dev).
#
# Requires: kubectl, helm, and a built k8mpatible binary.
#
# Usage:
#   ./e2e/run-e2e.sh <path-to-k8mpatible-binary>

set -euo pipefail

BINARY="${1:?Usage: $0 <k8mpatible-binary>}"
CLUSTER_NAME="k8mpatible-e2e"
PASS=0
FAIL=0

# Resolve kubeconfig from KIND
KUBECONFIG_PATH="$(mktemp)"
kind get kubeconfig --name "${CLUSTER_NAME}" > "${KUBECONFIG_PATH}"
export KUBECONFIG="${KUBECONFIG_PATH}"

cleanup() {
    rm -f "${KUBECONFIG_PATH}"
}
trap cleanup EXIT

echo "Using KIND cluster: ${CLUSTER_NAME}"
kubectl cluster-info
kubectl wait --for=condition=Ready nodes --all --timeout=120s

# ── Helper: run k8mpatible and capture output ──

run_k8mpatible() {
    local output_file
    output_file="$(mktemp)"
    echo "Running k8mpatible..."
    local exit_code=0
    "${BINARY}" --kubeconfig "${KUBECONFIG_PATH}" --output "${output_file}" 2>&1 || exit_code=$?
    echo "Exit code: ${exit_code}"
    echo "Output:"
    cat "${output_file}"
    echo ""
    # Return values via globals
    K8M_EXIT_CODE="${exit_code}"
    K8M_OUTPUT="$(cat "${output_file}")"
    rm -f "${output_file}"
}

# ── Assertion helpers ──

assert_exit_code() {
    local expected="$1"
    local test_name="$2"
    if [ "${K8M_EXIT_CODE}" -eq "${expected}" ]; then
        echo "PASS: ${test_name} (exit code ${K8M_EXIT_CODE})"
        PASS=$((PASS + 1))
    else
        echo "FAIL: ${test_name} - expected exit code ${expected}, got ${K8M_EXIT_CODE}"
        FAIL=$((FAIL + 1))
    fi
}

assert_output_contains() {
    local needle="$1"
    local test_name="$2"
    if echo "${K8M_OUTPUT}" | grep -qi "${needle}"; then
        echo "PASS: ${test_name}"
        PASS=$((PASS + 1))
    else
        echo "FAIL: ${test_name} - output does not contain '${needle}'"
        FAIL=$((FAIL + 1))
    fi
}

assert_output_not_contains() {
    local needle="$1"
    local test_name="$2"
    if echo "${K8M_OUTPUT}" | grep -qi "${needle}"; then
        echo "FAIL: ${test_name} - output unexpectedly contains '${needle}'"
        FAIL=$((FAIL + 1))
    else
        echo "PASS: ${test_name}"
        PASS=$((PASS + 1))
    fi
}

# ══════════════════════════════════════════════
# Install / uninstall helpers
# ══════════════════════════════════════════════

# Pre-add all Helm repos once to avoid repeated network calls in each test
setup_helm_repos() {
    echo "--- Adding Helm repositories ---"
    helm repo add jetstack https://charts.jetstack.io --force-update
    helm repo add istio https://istio-release.storage.googleapis.com/charts --force-update
    helm repo add kedacore https://kedacore.github.io/charts --force-update
    helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts --force-update
    helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts --force-update
    helm repo add jaegertracing https://jaegertracing.github.io/helm-charts --force-update
    helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts --force-update
    helm repo add kyverno https://kyverno.github.io/kyverno/ --force-update
    helm repo update
    echo "--- Helm repositories ready ---"
}

# ── cert-manager ──

install_cert_manager() {
    local version="$1"
    echo "--- Installing cert-manager ${version} ---"
    helm install cert-manager jetstack/cert-manager \
        --namespace cert-manager --create-namespace \
        --version "${version}" \
        --set crds.enabled=true \
        --set replicaCount=0 \
        --wait --timeout 120s
}

uninstall_cert_manager() {
    echo "--- Uninstalling cert-manager ---"
    helm uninstall cert-manager --namespace cert-manager --wait 2>/dev/null || true
    kubectl delete namespace cert-manager --wait=true 2>/dev/null || true
}

# ── Istio ──

install_istio() {
    local version="$1"
    echo "--- Installing Istio ${version} (istiod only) ---"
    helm install istio-base istio/base \
        --namespace istio-system --create-namespace \
        --version "${version}" \
        --wait --timeout 120s
    helm install istiod istio/istiod \
        --namespace istio-system \
        --version "${version}" \
        --wait --timeout 120s
}

uninstall_istio() {
    echo "--- Uninstalling Istio ---"
    helm uninstall istiod --namespace istio-system --wait 2>/dev/null || true
    helm uninstall istio-base --namespace istio-system --wait 2>/dev/null || true
    kubectl delete namespace istio-system --wait=true 2>/dev/null || true
}

# ── KEDA ──

install_keda() {
    local version="$1"
    echo "--- Installing KEDA ${version} ---"
    helm install keda kedacore/keda \
        --namespace keda --create-namespace \
        --version "${version}" \
        --set replicaCount=0 \
        --wait --timeout 120s
}

uninstall_keda() {
    echo "--- Uninstalling KEDA ---"
    helm uninstall keda --namespace keda --wait 2>/dev/null || true
    kubectl delete namespace keda --wait=true 2>/dev/null || true
}

# ── Velero ──

install_velero() {
    local version="$1"
    echo "--- Installing Velero ${version} ---"
    # Velero chart v11+ uses new configuration structure:
    # - configuration.provider removed, provider now per backupStorageLocation/volumeSnapshotLocation
    # - credentials.secretContents moved to credentials.secretContents.cloud
    helm install velero vmware-tanzu/velero \
        --namespace velero --create-namespace \
        --version "${version}" \
        --set configuration.backupStorageLocation[0].name=default \
        --set configuration.backupStorageLocation[0].provider=aws \
        --set configuration.backupStorageLocation[0].bucket=test-bucket \
        --set configuration.backupStorageLocation[0].config.region=us-east-1 \
        --set configuration.volumeSnapshotLocation[0].name=default \
        --set configuration.volumeSnapshotLocation[0].provider=aws \
        --set configuration.volumeSnapshotLocation[0].config.region=us-east-1 \
        --set initContainers[0].name=velero-plugin-for-aws \
        --set initContainers[0].image=velero/velero-plugin-for-aws:v1.12.0 \
        --set initContainers[0].volumeMounts[0].mountPath=/target \
        --set initContainers[0].volumeMounts[0].name=plugins \
        --set credentials.useSecret=false \
        --set replicaCount=0 \
        --wait --timeout 120s
}

uninstall_velero() {
    echo "--- Uninstalling Velero ---\"
    helm uninstall velero --namespace velero --wait 2>/dev/null || true
    kubectl delete namespace velero --wait=true 2>/dev/null || true
}

# ── Gatekeeper ──

install_gatekeeper() {
    local version="$1"
    echo "--- Installing Gatekeeper ${version} ---"
    helm install gatekeeper gatekeeper/gatekeeper \
        --namespace gatekeeper-system --create-namespace \
        --version "${version}" \
        --set replicaCount=0 \
        --wait --timeout 120s
}

uninstall_gatekeeper() {
    echo "--- Uninstalling Gatekeeper ---\"
    helm uninstall gatekeeper --namespace gatekeeper-system --wait 2>/dev/null || true
    kubectl delete namespace gatekeeper-system --wait=true 2>/dev/null || true
}

install_jaeger() {
    local version="$1"
    echo "--- Installing Jaeger Operator ${version} via Helm ---"
    helm install jaeger-operator jaegertracing/jaeger-operator \
        --namespace observability --create-namespace \
        --version "${version}" \
        --set crds.enabled=true \
        --set replicaCount=0 \
        --wait --timeout 120s
}

uninstall_jaeger() {
    echo "--- Uninstalling Jaeger ---"
    helm uninstall jaeger-operator --namespace observability --wait 2>/dev/null || true
    kubectl delete namespace observability --wait=true 2>/dev/null || true
}

install_opentelemetry() {
    local version="$1"
    echo "--- Installing OpenTelemetry Collector ${version} via Helm ---"
    helm install opentelemetry-collector open-telemetry/opentelemetry-collector \
        --namespace opentelemetry --create-namespace \
        --version "${version}" \
        --set mode=deployment \
        --set replicaCount=0 \
        --wait --timeout 120s
}

uninstall_opentelemetry() {
    echo "--- Uninstalling OpenTelemetry Collector ---"
    helm uninstall opentelemetry-collector --namespace opentelemetry --wait 2>/dev/null || true
    kubectl delete namespace opentelemetry --wait=true 2>/dev/null || true
}

# ── Kyverno (Helm-based, production-grade like a real user would use) ──

install_kyverno() {
    local version="$1"
    echo "--- Installing Kyverno ${version} via Helm ---"
    helm install kyverno kyverno/kyverno \
        --namespace kyverno --create-namespace \
        --version "${version}" \
        --set replicaCount=0 \
        --wait --timeout 120s
}

uninstall_kyverno() {
    echo "--- Uninstalling Kyverno ---"
    helm uninstall kyverno --namespace kyverno --wait 2>/dev/null || true
    kubectl delete namespace kyverno --wait=true 2>/dev/null || true
}

# ══════════════════════════════════════════════
# Test 1: Compatible versions
#   cert-manager 1.17.x on K8s 1.31 -> compatible (range >=1.29, <=1.32)
# ══════════════════════════════════════════════
test_compatible_versions() {
    echo ""
    echo "========================================="
    echo "TEST 1: Compatible tool versions"
    echo "========================================="

    # cert-manager v1.17.x is compatible with K8s 1.31
    install_cert_manager "v1.17.2"

    run_k8mpatible

    assert_exit_code 0 "Compatible cert-manager should produce exit code 0"
    assert_output_contains "cert-manager" "Output should list cert-manager as a discovered tool"
    assert_output_contains "Kubernetes" "Output should list Kubernetes"
    assert_output_not_contains "current_incompatibility" "No current incompatibilities expected"

    uninstall_cert_manager
}

# ══════════════════════════════════════════════
# Test 2: Incompatible version (too old for K8s)
#   Istio 1.17.x on K8s 1.31 -> incompatible (range >=1.23, <=1.26)
# ══════════════════════════════════════════════
test_incompatible_versions() {
    echo ""
    echo "========================================="
    echo "TEST 2: Incompatible tool version"
    echo "========================================="

    # Istio 1.17.x max K8s is 1.26 -- running on 1.31 is incompatible
    install_istio "1.17.8"

    run_k8mpatible

    assert_exit_code 1 "Incompatible Istio should produce exit code 1"
    assert_output_contains "Istio" "Output should list Istio as a discovered tool"

    uninstall_istio
}

# ══════════════════════════════════════════════
# Test 3: Mixed compatible + incompatible
#   cert-manager 1.17.x (compatible) + Istio 1.17.x (incompatible)
# ══════════════════════════════════════════════
test_mixed_compatibility() {
    echo ""
    echo "========================================="
    echo "TEST 3: Mixed compatible + incompatible"
    echo "========================================="

    install_cert_manager "v1.17.2"
    install_istio "1.17.8"

    run_k8mpatible

    assert_exit_code 1 "Mixed scenario should exit 1 due to incompatible Istio"
    assert_output_contains "cert-manager" "Output should list cert-manager"
    assert_output_contains "Istio" "Output should list Istio"

    uninstall_istio
    uninstall_cert_manager
}

# ══════════════════════════════════════════════
# Test 4: Tier-1 tool compatible — KEDA
#   KEDA 2.18.x on K8s 1.32 -> compatible (range >=1.31, <=1.33)
#   Must also survive upgrade simulation (1.32→1.33), so range must include 1.33
# ══════════════════════════════════════════════
test_keda_compatible() {
    echo ""
    echo "========================================="
    echo "TEST 4: KEDA compatible (tier-1 tool)"
    echo "========================================="

    install_keda "2.18.0"

    run_k8mpatible

    assert_exit_code 0 "Compatible KEDA should produce exit code 0"
    assert_output_contains "keda" "Output should list keda as a discovered tool"
    assert_output_not_contains "current_incompatibility" "No current incompatibilities expected"

    uninstall_keda
}

# ══════════════════════════════════════════════
# Test 5: Tier-1 tool incompatible — Kyverno (Helm)
#   Kyverno 1.12.x on K8s 1.31 -> incompatible (range >=1.26, <=1.29)
#   Uses Helm chart for production-grade installation
#   Helm chart v3.2.6 corresponds to Kyverno v1.12.x
# ══════════════════════════════════════════════
test_kyverno_incompatible() {
    echo ""
    echo "========================================="
    echo "TEST 5: Kyverno incompatible (tier-1 tool)"
    echo "========================================="

    install_kyverno "3.2.6"

    run_k8mpatible

    assert_exit_code 1 "Incompatible Kyverno should produce exit code 1"
    assert_output_contains "kyverno" "Output should list kyverno as a discovered tool"

    uninstall_kyverno
}

# ══════════════════════════════════════════════
# Test 6: Mixed tier-1 tools (compatible + incompatible)
#   KEDA 2.18.x (compatible, range >=1.31, <=1.33)
#   + Kyverno 1.12.x (incompatible, range >=1.26, <=1.29)
#   Helm chart v3.2.6 corresponds to Kyverno v1.12.x
# ══════════════════════════════════════════════
test_mixed_tier1() {
    echo ""
    echo "========================================="
    echo "TEST 6: Mixed tier-1 tools (KEDA compatible + Kyverno incompatible)"
    echo "========================================="

    install_keda "2.18.0"
    install_kyverno "3.2.6"

    run_k8mpatible

    assert_exit_code 1 "Mixed tier-1 should exit 1 due to incompatible Kyverno"
    assert_output_contains "keda" "Output should list keda"
    assert_output_contains "kyverno" "Output should list kyverno"

    uninstall_kyverno
    uninstall_keda
}

# ══════════════════════════════════════════════
# Test 7: Tier-2 tool compatible — Velero
#   Velero 1.16 on K8s 1.31 -> compatible (range >=1.18, <=1.33)
#   Must also survive upgrade simulation (1.31→1.32), range includes 1.32
# ══════════════════════════════════════════════
test_velero_compatible() {
    echo ""
    echo "========================================="
    echo "TEST 7: Velero compatible (tier-2 tool)"
    echo "========================================="

    install_velero "11.0.0"

    run_k8mpatible

    assert_exit_code 0 "Compatible Velero should produce exit code 0"
    assert_output_contains "velero" "Output should list velero as a discovered tool"
    assert_output_not_contains "current_incompatibility" "No current incompatibilities expected"

    uninstall_velero
}

# ══════════════════════════════════════════════
# Test 8: Tier-2 tool incompatible — Gatekeeper
#   Gatekeeper 3.14 on K8s 1.31 -> incompatible (range >=1.22, <=1.28)
# ══════════════════════════════════════════════
test_gatekeeper_incompatible() {
    echo ""
    echo "========================================="
    echo "TEST 8: Gatekeeper incompatible (tier-2 tool)"
    echo "========================================="

    install_gatekeeper "3.14.2"

    run_k8mpatible

    assert_exit_code 1 "Incompatible Gatekeeper should produce exit code 1"
    assert_output_contains "gatekeeper" "Output should list gatekeeper as a discovered tool"

    uninstall_gatekeeper
}

# ══════════════════════════════════════════════
# Test 9: Tier-2 tool incompatible — Jaeger (Helm)
#   Jaeger Operator 2.x supports K8s 1.29–1.32 (per chart)
#   On K8s 1.31 -> compatible, but we test older version 1.62 equivalent
#   Jaeger Operator v2.57.0 chart version corresponds to Jaeger v1.62+
#   Using older chart version that maps to Jaeger 1.62 (incompatible with 1.31)
# ══════════════════════════════════════════════
test_jaeger_incompatible() {
    echo ""
    echo "========================================="
    echo "TEST 9: Jaeger incompatible (tier-2 tool, Helm)"
    echo "========================================="

    # Jaeger Operator v2.42.0 chart maps to Jaeger ~1.53, incompatible with K8s 1.31
    install_jaeger "v2.42.0"

    run_k8mpatible

    assert_exit_code 1 "Incompatible Jaeger should produce exit code 1"
    assert_output_contains "jaeger" "Output should list jaeger as a discovered tool"

    uninstall_jaeger
}

# ══════════════════════════════════════════════
# Test 10: Tier-2 tool compatible — OpenTelemetry Collector (Helm)
#   OpenTelemetry Collector 0.129 (range >=1.23, <=1.33)
#   On K8s 1.31 -> compatible (includes 1.31 and upgrade target 1.32)
#   Chart version 0.129.0 maps to collector 0.129.x
# ══════════════════════════════════════════════
test_opentelemetry_compatible() {
    echo ""
    echo "========================================="
    echo "TEST 10: OpenTelemetry Collector compatible (tier-2 tool, Helm)"
    echo "========================================="

    install_opentelemetry "0.129.0"

    run_k8mpatible

    assert_exit_code 0 "Compatible OpenTelemetry Collector should produce exit code 0"
    assert_output_contains "opentelemetry" "Output should list opentelemetry as a discovered tool"
    assert_output_not_contains "current_incompatibility" "No current incompatibilities expected"

    uninstall_opentelemetry
}

# ══════════════════════════════════════════════
# Test 11: Mixed tier-2 tools (compatible + incompatible)
#   Velero 1.16 (compatible, range >=1.18, <=1.33)
#   + Gatekeeper 3.14 (incompatible, range >=1.22, <=1.28)
# ══════════════════════════════════════════════
test_mixed_tier2() {
    echo ""
    echo "========================================="
    echo "TEST 11: Mixed tier-2 tools (Velero compatible + Gatekeeper incompatible)"
    echo "========================================="

    install_velero "11.0.0"
    install_gatekeeper "3.14.2"

    run_k8mpatible

    assert_exit_code 1 "Mixed tier-2 should exit 1 due to incompatible Gatekeeper"
    assert_output_contains "velero" "Output should list velero"
    assert_output_contains "gatekeeper" "Output should list gatekeeper"

    uninstall_gatekeeper
    uninstall_velero
}

# ══════════════════════════════════════════════
# Run all tests
# ══════════════════════════════════════════════
main() {
    # Pre-add all Helm repos once to avoid repeated network calls
    setup_helm_repos

    test_compatible_versions
    test_incompatible_versions
    test_mixed_compatibility
    test_keda_compatible
    test_kyverno_incompatible
    test_mixed_tier1
    test_velero_compatible
    test_gatekeeper_incompatible
    test_jaeger_incompatible
    test_opentelemetry_compatible
    test_mixed_tier2

    echo ""
    echo "========================================="
    echo "E2E Results: ${PASS} passed, ${FAIL} failed"
    echo "========================================="

    if [ "${FAIL}" -gt 0 ]; then
        exit 1
    fi
}

main
