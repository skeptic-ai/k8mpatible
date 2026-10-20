package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Exercise the shipped data, not only the synthetic graph fixture.
func TestCompatibilityFiles(t *testing.T) {
	files, err := filepath.Glob("compatibility/*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 23 {
		t.Fatalf("expected all 23 compatibility definitions, got %d", len(files))
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			graph, err := LoadGraphFromYAML(file)
			if err != nil {
				t.Fatal(err)
			}
			if len(graph.Nodes) == 0 {
				t.Fatal("compatibility definition must declare a node")
			}
			nodes := map[string]bool{}
			for _, node := range graph.Nodes {
				if node.Name == "" || nodes[node.Name] {
					t.Fatalf("empty or duplicate node name: %q", node.Name)
				}
				nodes[node.Name] = true
				u, err := url.Parse(node.DocUrl)
				if err != nil || u == nil || u.Scheme != "https" || u.Host == "" {
					t.Errorf("%s: invalid documentation URL %q", node.Name, node.DocUrl)
				}
			}
			positive := map[string]bool{}
			for _, edge := range graph.Edges {
				if !nodes[edge.SourceName] || edge.DestinationName == "" || edge.SourceVersion == nil || edge.DestinationVersionRange == nil {
					t.Fatalf("incomplete or undeclared-source edge: %+v", edge)
				}
				if edge.Compatible {
					key := edge.SourceName + "/" + edge.SourceVersion.String() + "/" + edge.DestinationName
					if positive[key] {
						t.Errorf("duplicate positive edge: %s", key)
					}
					positive[key] = true
				}
			}
		})
	}
}

func TestOpenTelemetryComponentIdentity(t *testing.T) {
	graph, err := LoadGraphFromYAML("compatibility/opentelemetry-collector.yaml")
	if err != nil {
		t.Fatal(err)
	}
	// A local API fixture tests discovery wiring, not real runtime compatibility.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apis/apps/v1/deployments" {
			http.NotFound(w, r)
			return
		}
		name, version, image := "opentelemetry-collector", "0.163.0", "otel/opentelemetry-collector-contrib"
		if r.URL.Query().Get("labelSelector") == "app.kubernetes.io/name=opentelemetry-operator" {
			name, version, image = "opentelemetry-operator", "0.161.0", "ghcr.io/open-telemetry/opentelemetry-operator/opentelemetry-operator"
		}
		w.Header().Set("Content-Type", "application/json")
		list := appsv1.DeploymentList{Items: []appsv1.Deployment{{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: name, Image: image + ":" + version}},
			}}},
		}}}
		if err := json.NewEncoder(w).Encode(list); err != nil {
			t.Errorf("encode API fixture: %v", err)
		}
	}))
	defer server.Close()
	clientset, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	tools := discoverTools(clientset, graph)
	if len(tools) != 2 {
		t.Fatalf("expected separate Collector and Operator discovery, got %+v", tools)
	}
	for name, expected := range map[string]string{"opentelemetry-collector": "0.163.0", "opentelemetry-operator": "0.161.0"} {
		version := GetActiveVersion(tools, name)
		if version == nil || version.String() != expected {
			t.Errorf("%s: expected %s, got %v", name, expected, version)
		}
	}
	got, _ := graph.CheckCompatibility("opentelemetry-operator", semver.MustParse("0.161.0"), "cert-manager", semver.MustParse("1.21.2"))
	if got != Compatible {
		t.Error("Operator should retain its documented cert-manager v1 edge")
	}
	got, _ = graph.CheckCompatibility("opentelemetry-operator", semver.MustParse("0.161.0"), "cert-manager", semver.MustParse("2.0.0"))
	if got != NotCompatible {
		t.Error("Operator's cert-manager v1 envelope must reject cert-manager v2")
	}
	got, _ = graph.CheckCompatibility("opentelemetry-operator", semver.MustParse("0.161.0"), "prometheus-operator", semver.MustParse("0.94.1"))
	if got != Unknown {
		t.Error("Operator must not invent a prometheus-operator compatibility edge")
	}
	got, _ = graph.CheckCompatibility("opentelemetry-collector", semver.MustParse("0.163.0"), "cert-manager", semver.MustParse("1.21.2"))
	if got != Unknown {
		t.Error("standalone Collector must not inherit an Operator dependency")
	}
}

// NotCompatible outside a positive envelope is the engine's result, not proof
// that upstream demonstrated breakage on those Kubernetes versions.
func TestLatestReleaseKubernetesCoverage(t *testing.T) {
	for _, tc := range []struct {
		file, tool, version string
		want135, want136    int
	}{
		{"argocd", "argocd", "3.5.5", Compatible, Compatible},
		{"argoworkflows", "argoworkflows", "4.1.5", Compatible, Compatible},
		{"aws-lb-controller", "aws-load-balancer-controller", "3.6.0", Compatible, Compatible},
		{"cert-manager", "cert-manager", "1.21.2", Compatible, Compatible},
		{"cilium", "cilium", "1.20.3", Compatible, Compatible},
		{"datadog", "datadog-agent", "7.84.4", Compatible, Compatible},
		{"datadog", "datadog-cluster-agent", "7.84.4", Compatible, Compatible},
		{"external-dns", "external-dns", "0.23.0", Compatible, Compatible},
		{"external-secrets", "external-secrets", "2.13.0", NotCompatible, Compatible},
		{"falco", "falco", "0.45.1", Unknown, Unknown},
		{"flux", "flux", "2.9.6", Compatible, Compatible},
		{"gatekeeper", "gatekeeper", "3.23.1", Compatible, Compatible},
		{"istio", "Istio", "1.31.2", Compatible, Compatible},
		{"jaeger", "jaeger", "2.22.0", Unknown, Unknown},
		{"keda", "keda", "2.21.0", Compatible, Compatible},
		{"kyverno", "kyverno", "1.19.1", Compatible, NotCompatible},
		{"linkerd", "Linkerd", "2.20.0", Compatible, NotCompatible},
		{"nginx-ingress", "nginx-ingress-controller", "5.7.0", Compatible, Compatible},
		{"opentelemetry-collector", "opentelemetry-operator", "0.161.0", Compatible, Compatible},
		{"opentelemetry-collector", "opentelemetry-collector", "0.163.0", Unknown, Unknown},
		{"prometheus", "prometheus-operator", "0.94.1", Compatible, Compatible},
		{"stakater-reloader", "stakater-reloader", "1.4.22", Compatible, Compatible},
		{"vault-secrets-operator", "vault-secrets-operator", "1.6.1", Compatible, Compatible},
		{"velero", "velero", "1.18.4", Compatible, NotCompatible},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			graph, err := LoadGraphFromYAML("compatibility/" + tc.file + ".yaml")
			if err != nil {
				t.Fatal(err)
			}
			declared := false
			for _, node := range graph.Nodes {
				declared = declared || node.Name == tc.tool
			}
			if !declared {
				t.Errorf("missing node %s", tc.tool)
			}
			for _, target := range []struct {
				version string
				want    int
			}{{"1.35.9", tc.want135}, {"1.36.5", tc.want136}} {
				got, reason := graph.CheckCompatibility(tc.tool, semver.MustParse(tc.version), Kubernetes, semver.MustParse(target.version))
				if got != target.want {
					t.Errorf("Kubernetes %s: got %d (%s), want %d", target.version, got, reason, target.want)
				}
			}
		})
	}
}

func TestNGINXKubernetes135And136(t *testing.T) {
	graph, err := LoadGraphFromYAML("compatibility/nginx-ingress.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		tool, kubernetes string
		want             int
	}{
		{"5.7.0", "1.35.9", Compatible},
		{"5.7.0", "1.36.5", Compatible},
		{"5.6.3", "1.36.5", Compatible},
		{"5.5.4", "1.36.5", Compatible},
		{"5.5.4", "1.37.0", NotCompatible},
		{"5.4.3", "1.35.9", Compatible},
		{"5.4.3", "1.36.0", NotCompatible},
	} {
		t.Run(tc.tool+"/"+tc.kubernetes, func(t *testing.T) {
			got, reason := graph.CheckCompatibility("nginx-ingress-controller", semver.MustParse(tc.tool), Kubernetes, semver.MustParse(tc.kubernetes))
			if got != tc.want {
				t.Errorf("got %d (%s), want %d", got, reason, tc.want)
			}
		})
	}
	// Check the raw constraint as well: callers must not need to strip patches.
	for _, edge := range graph.Edges {
		if edge.SourceVersion.Equal(semver.MustParse("5.5.0")) && !edge.DestinationVersionRange.Check(semver.MustParse("1.36.5")) {
			t.Error("NGINX 5.5 constraint must include Kubernetes 1.36 patch releases")
		}
	}
}
