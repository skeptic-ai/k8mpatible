package client

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestCertManagerEKSProviderVersion(t *testing.T) {
	graph, err := LoadGraphFromYAML("compatibility/cert-manager.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		version string
		want    int
	}{
		{"1.34.6", Compatible},
		{"1.34.6-eksbuild.49", Compatible},
		{"1.34.6-eks-abcd123", Compatible},
		{"1.34.6-gke.1234000", Compatible},
		{"1.34.6+k3s1", Compatible},
		{"1.36.0-eksbuild.49", NotCompatible},
		{"1.31.9-eksbuild.49", NotCompatible},
		{"1.34.6-rc.1", NotCompatible},
		{"1.34.6-beta.1", NotCompatible},
		{"1.34.6-alpha.1", NotCompatible},
		{"1.34.6-custombuild.49", NotCompatible},
		{"1.34.6-rc.1-eksbuild.49", NotCompatible},
	} {
		t.Run(tc.version, func(t *testing.T) {
			version := semver.MustParse(tc.version)
			before := version.String()
			got, reason := graph.CheckCompatibility("cert-manager", semver.MustParse("1.20.4"), Kubernetes, version)
			if got != tc.want {
				t.Errorf("cert-manager 1.20.4 on Kubernetes %s: got %d (%s), want %d", tc.version, got, reason, tc.want)
			}
			if version.String() != before {
				t.Error("comparison must not mutate the discovered version")
			}
		})
	}
}

func TestProviderSuffixMinimumConstraint(t *testing.T) {
	constraint, err := semver.NewConstraint(">=1.22")
	if err != nil {
		t.Fatal(err)
	}
	graph := &Graph{Edges: []*Edge{{
		SourceName: "fixture", SourceVersion: semver.MustParse("1.0.0"),
		DestinationName: Kubernetes, DestinationVersionRange: constraint, Compatible: true,
	}}}
	version := semver.MustParse("1.34.6-eksbuild.49")
	_, diagnostics := constraint.Validate(version)
	t.Logf("raw semver prerelease=%q, validation diagnostics=%v", version.Prerelease(), diagnostics)
	got, reason := graph.CheckCompatibility("fixture", semver.MustParse("1.0.0"), Kubernetes, version)
	if got != Compatible {
		t.Fatalf("provider build satisfies >=1.22: got %d (%s), want Compatible", got, reason)
	}
}

func TestProviderVersionRetainsPatchMinimum(t *testing.T) {
	graph, err := LoadGraphFromYAML("compatibility/flux.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		version string
		want    int
	}{
		{"1.34.0-eksbuild.49", NotCompatible},
		{"1.34.1-eksbuild.49", Compatible},
		{"1.34.6-eksbuild.49", Compatible},
		{"1.34.1-gke.1234000", Compatible},
	} {
		t.Run(tc.version, func(t *testing.T) {
			got, reason := graph.CheckCompatibility("flux", semver.MustParse("2.9.6"), Kubernetes, semver.MustParse(tc.version))
			if got != tc.want {
				t.Errorf("got %d (%s), want %d", got, reason, tc.want)
			}
		})
	}
}

func TestNonKubernetesPrereleaseRemainsUnchanged(t *testing.T) {
	for _, constraint := range []string{">=1.22", ">=1.34.0-0, <1.35"} {
		t.Run(constraint, func(t *testing.T) {
			rangeConstraint, err := semver.NewConstraint(constraint)
			if err != nil {
				t.Fatal(err)
			}
			graph := &Graph{Edges: []*Edge{{
				SourceName: "fixture", SourceVersion: semver.MustParse("1.0.0"),
				DestinationName: "other-tool", DestinationVersionRange: rangeConstraint, Compatible: true,
			}}}
			version := semver.MustParse("1.34.6-eksbuild.49")
			want := NotCompatible
			if rangeConstraint.Check(version) {
				want = Compatible
			}
			got, reason := graph.CheckCompatibility("fixture", semver.MustParse("1.0.0"), "other-tool", version)
			if got != want {
				t.Errorf("non-Kubernetes semver behavior changed: got %d (%s), want %d", got, reason, want)
			}
		})
	}
}
