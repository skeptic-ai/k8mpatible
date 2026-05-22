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

# ── Kyverno (kubectl-based, avoiding Helm chart version mismatch) ──

install_kyverno_kubectl() {
    local app_version="$1"
    echo "--- Installing Kyverno ${app_version} via kubectl ---"
    kubectl create namespace kyverno --dry-run=client -o yaml | kubectl apply -f -
    kubectl create deployment kyverno-admission-controller \
        --namespace kyverno \
        --image "nginx:v1.12.6" \
        --dry-run=client -o yaml | kubectl apply -f -
    # Add required labels for k8mpatible discovery
    kubectl label deployment kyverno-admission-controller \
        --namespace kyverno \
        app.kubernetes.io/name=kyverno \
        app.kubernetes.io/component=admission-controller \
        --overwrite
    kubectl wait --for=condition=Available deployment/kyverno-admission-controller \
        --namespace kyverno --timeout=120s
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
#   KEDA 2.17.x on K8s 1.31 -> compatible (range >=1.30, <=1.32)
#   Must also survive upgrade simulation (1.31→1.32), so range must include 1.32
# ══════════════════════════════════════════════
test_keda_compatible() {
    echo ""
    echo "========================================="
    echo "TEST 4: KEDA compatible (tier-1 tool)"
    echo "========================================="

    install_keda "2.17.0"

    run_k8mpatible

    assert_exit_code 0 "Compatible KEDA should produce exit code 0"
    assert_output_contains "keda" "Output should list keda as a discovered tool"
    assert_output_not_contains "current_incompatibility" "No current incompatibilities expected"

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
#   KEDA 2.17.x (compatible, range >=1.30, <=1.32)
#   + Kyverno 1.12.x (incompatible, range >=1.26, <=1.29)
# ══════════════════════════════════════════════
test_mixed_tier1() {
    echo ""
    echo "========================================="
    echo "TEST 6: Mixed tier-1 tools (KEDA compatible + Kyverno incompatible)"
    echo "========================================="

    install_keda "2.17.0"
    install_kyverno_kubectl "v1.12.6"

    run_k8mpatible

    assert_exit_code 1 "Mixed tier-1 should exit 1 due to incompatible Kyverno"
    assert_output_contains "keda" "Output should list keda"
    assert_output_contains "kyverno" "Output should list kyverno"

    uninstall_kyverno_kubectl
    uninstall_keda
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

    echo ""
    echo "========================================="
    echo "E2E Results: ${PASS} passed, ${FAIL} failed"
    echo "========================================="

    if [ "${FAIL}" -gt 0 ]; then
        exit 1
    fi
}

main
