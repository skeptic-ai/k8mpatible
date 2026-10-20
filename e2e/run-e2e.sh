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

# ── cert-manager ──

install_cert_manager() {
    local version="$1"
    echo "--- Installing cert-manager ${version} ---"
    helm repo add jetstack https://charts.jetstack.io --force-update
    helm repo update jetstack
    helm install cert-manager jetstack/cert-manager \
        --namespace cert-manager --create-namespace \
        --version "${version}" \
        --set crds.enabled=true \
        --wait --timeout 180s
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
    helm repo add istio https://istio-release.storage.googleapis.com/charts --force-update
    helm repo update istio
    helm install istio-base istio/base \
        --namespace istio-system --create-namespace \
        --version "${version}" \
        --wait --timeout 120s
    helm install istiod istio/istiod \
        --namespace istio-system \
        --version "${version}" \
        --wait --timeout 180s
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
    helm repo add kedacore https://kedacore.github.io/charts --force-update
    helm repo update kedacore
    helm install keda kedacore/keda \
        --namespace keda --create-namespace \
        --version "${version}" \
        --wait --timeout 180s
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
    helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts --force-update
    helm repo update vmware-tanzu
    helm install velero vmware-tanzu/velero \
        --namespace velero --create-namespace \
        --version "${version}" \
        --set configuration.backupStorageLocation[0].name=default \
        --set configuration.backupStorageLocation[0].provider=aws \
        --set configuration.backupStorageLocation[0].bucket=velero-test \
        --set configuration.backupStorageLocation[0].config.region=us-east-1 \
        --set configuration.volumeSnapshotLocation[0].name=default \
        --set configuration.volumeSnapshotLocation[0].provider=aws \
        --set configuration.volumeSnapshotLocation[0].config.region=us-east-1 \
        --set initContainers[0].name=velero-plugin-for-aws \
        --set initContainers[0].image=velero/velero-plugin-for-aws:v1.12.0 \
        --set initContainers[0].volumeMounts[0].mountPath=/target \
        --set initContainers[0].volumeMounts[0].name=plugins \
        --set credentials.useSecret=false \
        --wait --timeout 180s
}

uninstall_velero() {
    echo "--- Uninstalling Velero ---"
    helm uninstall velero --namespace velero --wait 2>/dev/null || true
    kubectl delete namespace velero --wait=true 2>/dev/null || true
}

# ── Gatekeeper ──

install_gatekeeper() {
    local version="$1"
    echo "--- Installing Gatekeeper ${version} ---"
    helm repo add gatekeeper https://open-policy-agent.github.io/gatekeeper/charts --force-update
    helm repo update gatekeeper
    helm install gatekeeper gatekeeper/gatekeeper \
        --namespace gatekeeper-system --create-namespace \
        --version "${version}" \
        --wait --timeout 180s
}

uninstall_gatekeeper() {
    echo "--- Uninstalling Gatekeeper ---"
    helm uninstall gatekeeper --namespace gatekeeper-system --wait 2>/dev/null || true
    kubectl delete namespace gatekeeper-system --wait=true 2>/dev/null || true
}

# ── Jaeger (kubectl-based, similar to Kyverno approach) ──

install_jaeger_kubectl() {
    local app_version="$1"
    echo "--- Installing Jaeger ${app_version} via kubectl ---"
    kubectl create namespace observability --dry-run=client -o yaml | kubectl apply -f -
    kubectl create deployment jaeger \
        --namespace observability \
        --image "jaegertracing/jaeger:${app_version}" \
        --dry-run=client -o yaml | kubectl apply -f -
    # Add required labels for k8mpatible discovery
    kubectl label deployment jaeger \
        --namespace observability \
        app=jaeger \
        app.kubernetes.io/name=jaeger \
        --overwrite
    # No kubectl wait — deployment only needs to exist for discovery, not be Available
    sleep 5
}

uninstall_jaeger_kubectl() {
    echo "--- Uninstalling Jaeger ---"
    kubectl delete deployment jaeger --namespace observability --wait 2>/dev/null || true
    kubectl delete namespace observability --wait=true 2>/dev/null || true
}

# ── OpenTelemetry Collector (kubectl-based) ──

install_opentelemetry_kubectl() {
    local app_version="$1"
    echo "--- Installing OpenTelemetry Collector ${app_version} via kubectl ---"
    kubectl create namespace opentelemetry --dry-run=client -o yaml | kubectl apply -f -
    kubectl create deployment opentelemetry-collector \
        --namespace opentelemetry \
        --image "otel/opentelemetry-collector-contrib:${app_version}" \
        --dry-run=client -o yaml | kubectl apply -f -
    # Add required labels for k8mpatible discovery
    kubectl label deployment opentelemetry-collector \
        --namespace opentelemetry \
        app.kubernetes.io/name=opentelemetry-collector \
        --overwrite
    sleep 5
}

uninstall_opentelemetry_kubectl() {
    echo "--- Uninstalling OpenTelemetry Collector ---"
    kubectl delete deployment opentelemetry-collector --namespace opentelemetry --wait 2>/dev/null || true
    kubectl delete namespace opentelemetry --wait=true 2>/dev/null || true
}

# Operator fixture: discovery/integration assertions, NOT a running installation.
# No CRDs, admission certificates, or ready controller are claimed by this helper.
install_opentelemetry_operator_fixture() {
    local app_version="$1"
    kubectl create namespace opentelemetry-operator-system --dry-run=client -o yaml | kubectl apply -f -
    kubectl create deployment opentelemetry-operator-controller-manager \
        --namespace opentelemetry-operator-system \
        --image "ghcr.io/open-telemetry/opentelemetry-operator/opentelemetry-operator:${app_version}" \
        --dry-run=client -o yaml | kubectl apply -f -
    kubectl label deployment opentelemetry-operator-controller-manager \
        --namespace opentelemetry-operator-system \
        app.kubernetes.io/name=opentelemetry-operator --overwrite
}

uninstall_opentelemetry_operator_fixture() {
    kubectl delete namespace opentelemetry-operator-system --wait=true 2>/dev/null || true
}

# ── Kyverno (kubectl-based, avoiding Helm chart version mismatch) ──

install_kyverno_kubectl() {
    local app_version="$1"
    echo "--- Installing Kyverno ${app_version} via kubectl ---"
    kubectl create namespace kyverno --dry-run=client -o yaml | kubectl apply -f -
    kubectl create deployment kyverno-admission-controller \
        --namespace kyverno \
        --image "ghcr.io/kyverno/kyverno:${app_version}" \
        --dry-run=client -o yaml | kubectl apply -f -
    # Add required labels for k8mpatible discovery
    kubectl label deployment kyverno-admission-controller \
        --namespace kyverno \
        app.kubernetes.io/name=kyverno \
        app.kubernetes.io/component=admission-controller \
        --overwrite
    # No kubectl wait — deployment only needs to exist for discovery, not be Available
    sleep 5
}

uninstall_kyverno_kubectl() {
    echo "--- Uninstalling Kyverno ---"
    kubectl delete deployment kyverno-admission-controller --namespace kyverno --wait 2>/dev/null || true
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
    assert_output_contains "^[[:space:]]*incompatibleTools: 0$" "No current incompatibilities expected"

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
    assert_output_contains "^[[:space:]]*incompatibleTools: 0$" "No current incompatibilities expected"

    uninstall_keda
}

# ══════════════════════════════════════════════
# Test 5: Tier-1 tool incompatible — Kyverno (kubectl-based)
#   Kyverno 1.12.x on K8s 1.31 -> incompatible (range >=1.26, <=1.29)
#   Uses kubectl create deployment to avoid Helm chart version mismatch
# ══════════════════════════════════════════════
test_kyverno_incompatible() {
    echo ""
    echo "========================================="
    echo "TEST 5: Kyverno incompatible (tier-1 tool)"
    echo "========================================="

    install_kyverno_kubectl "v1.12.6"

    run_k8mpatible

    assert_exit_code 1 "Incompatible Kyverno should produce exit code 1"
    assert_output_contains "kyverno" "Output should list kyverno as a discovered tool"

    uninstall_kyverno_kubectl
}

# ══════════════════════════════════════════════
# Test 6: Mixed tier-1 tools (compatible + incompatible)
#   KEDA 2.18.x (compatible, range >=1.31, <=1.33)
#   + Kyverno 1.12.x (incompatible, range >=1.26, <=1.29)
# ══════════════════════════════════════════════
test_mixed_tier1() {
    echo ""
    echo "========================================="
    echo "TEST 6: Mixed tier-1 tools (KEDA compatible + Kyverno incompatible)"
    echo "========================================="

    install_keda "2.18.0"
    install_kyverno_kubectl "v1.12.6"

    run_k8mpatible

    assert_exit_code 1 "Mixed tier-1 should exit 1 due to incompatible Kyverno"
    assert_output_contains "keda" "Output should list keda"
    assert_output_contains "kyverno" "Output should list kyverno"

    uninstall_kyverno_kubectl
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
    assert_output_contains "^[[:space:]]*incompatibleTools: 0$" "No current incompatibilities expected"

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
# Test 9: Jaeger runtime discovery with unknown Kubernetes compatibility.
#   The runtime is not the retired Jaeger Operator; do not apply its matrix.
#   Unknown currently does not count as an incompatibility in CLI exit status.
# ══════════════════════════════════════════════
test_jaeger_unknown() {
    echo ""
    echo "========================================="
    echo "TEST 9: Jaeger runtime unknown (discovery fixture)"
    echo "========================================="

    install_jaeger_kubectl "2.22.0"

    run_k8mpatible

    assert_exit_code 0 "Unknown Jaeger runtime should not invent incompatibility"
    assert_output_contains "^[[:space:]]*incompatibleTools: 0$" "No operator-derived runtime incompatibility expected"
    assert_output_contains "jaeger" "Output should list jaeger as a discovered tool"

    uninstall_jaeger_kubectl
}

# ══════════════════════════════════════════════
# Test 10: OpenTelemetry Collector discovery with unknown runtime compatibility.
#   The Operator matrix must not be applied to this standalone runtime.
# ══════════════════════════════════════════════
test_opentelemetry_unknown() {
    echo ""
    echo "========================================="
    echo "TEST 10: OpenTelemetry Collector unknown (discovery fixture)"
    echo "========================================="

    install_opentelemetry_kubectl "0.129.0"

    run_k8mpatible

    assert_exit_code 0 "Unknown Collector runtime should not invent incompatibility"
    assert_output_contains "opentelemetry" "Output should list opentelemetry as a discovered tool"
    assert_output_contains "^[[:space:]]*incompatibleTools: 0$" "No current incompatibilities expected"

    uninstall_opentelemetry_kubectl
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
# Test 12: OpenTelemetry Operator + cert-manager co-discovery
#   The scanner's lexical ordering checks cert-manager -> Operator, not the
#   reverse Operator -> cert-manager dependency; direct graph tests cover that edge.
#   cert-manager 1.17 is compatible with K8s 1.32
#   Both should be discovered with no known incompatibilities; this does not
#   validate the Operator dependency or a running webhook installation.
# ══════════════════════════════════════════════
test_opentelemetry_operator_codiscovery() {
    echo ""
    echo "========================================="
    echo "TEST 12: Co-discovery (OTel Operator + cert-manager)"
    echo "========================================="

    install_cert_manager "v1.17.2"
    install_opentelemetry_operator_fixture "0.129.1"

    run_k8mpatible

    assert_exit_code 0 "OTel Operator + cert-manager co-discovery should produce exit code 0"
    assert_output_contains "opentelemetry" "Output should list opentelemetry as a discovered tool"
    assert_output_contains "cert-manager" "Output should list cert-manager as a discovered tool"
    assert_output_contains "^[[:space:]]*incompatibleTools: 0$" "No current incompatibilities expected between OTel and cert-manager"

    uninstall_opentelemetry_operator_fixture
    uninstall_cert_manager
}

# ══════════════════════════════════════════════
# Run all tests
# ══════════════════════════════════════════════
main() {
    test_compatible_versions
    test_incompatible_versions
    test_mixed_compatibility
    test_keda_compatible
    test_kyverno_incompatible
    test_mixed_tier1
    test_velero_compatible
    test_gatekeeper_incompatible
    test_jaeger_unknown
    test_opentelemetry_unknown
    test_mixed_tier2
    test_opentelemetry_operator_codiscovery

    echo ""
    echo "========================================="
    echo "E2E Results: ${PASS} passed, ${FAIL} failed"
    echo "========================================="

    if [ "${FAIL}" -gt 0 ]; then
        exit 1
    fi
}

main
