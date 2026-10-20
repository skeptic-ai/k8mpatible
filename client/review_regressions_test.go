package client

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestCheckCompatibilityDestinationPatchMinimum(t *testing.T) {
	for _, destination := range []struct {
		name, constraint, minimum, supportedPatch, belowMinimum, excludedMinor string
	}{
		{Kubernetes, ">=1.34.1, <1.35", "1.34.1", "1.34.9", "1.34.0", "1.35.0"},
		{"Prometheus", ">=2.25.1, <2.26", "2.25.1", "2.25.9", "2.25.0", "2.26.0"},
	} {
		t.Run(destination.name, func(t *testing.T) {
			constraint, err := semver.NewConstraint(destination.constraint)
			if err != nil {
				t.Fatal(err)
			}
			graph := &Graph{Edges: []*Edge{{
				SourceName:              "source",
				SourceVersion:           semver.MustParse("1.0.0"),
				DestinationName:         destination.name,
				DestinationVersionRange: constraint,
				Compatible:              true,
			}}}
			for _, tc := range []struct {
				version string
				want    int
			}{
				{destination.belowMinimum, NotCompatible},
				{destination.minimum, Compatible},
				{destination.supportedPatch, Compatible},
				{destination.excludedMinor, NotCompatible},
			} {
				t.Run(tc.version, func(t *testing.T) {
					got, reason := graph.CheckCompatibility("source", semver.MustParse("1.0.7"), destination.name, semver.MustParse(tc.version))
					if got != tc.want {
						t.Errorf("got %d (%s), want %d", got, reason, tc.want)
					}
				})
			}
		})
	}
}

func TestGatekeeperMaintainedKubernetes134(t *testing.T) {
	graph, err := LoadGraphFromYAML("compatibility/gatekeeper.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"3.22.0", "3.23.1"} {
		for _, destination := range []struct {
			version string
			want    int
		}{
			{"1.30.9", NotCompatible},
			{"1.31.6", Compatible},
			{"1.33.2", Compatible},
			{"1.34.0", Compatible},
			{"1.34.12", Compatible},
			{"1.35.9", Compatible},
			{"1.36.5", Compatible},
			{"1.37.1", Compatible},
			{"1.38.0", NotCompatible},
		} {
			t.Run(source+"/"+destination.version, func(t *testing.T) {
				got, reason := graph.CheckCompatibility("gatekeeper", semver.MustParse(source), Kubernetes, semver.MustParse(destination.version))
				if got != destination.want {
					t.Errorf("got %d (%s), want %d", got, reason, destination.want)
				}
			})
		}
	}
}

func TestE2EOpenTelemetryFixtureVersions(t *testing.T) {
	script, err := os.ReadFile("../e2e/run-e2e.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, call := range []string{
		`install_opentelemetry_operator_fixture "0.129.1"`,
		`install_opentelemetry_kubectl "0.129.0"`,
	} {
		if !strings.Contains(string(script), call) {
			t.Errorf("missing released component fixture: %s", call)
		}
	}
}

func TestE2EZeroIncompatibilityAssertions(t *testing.T) {
	script, err := os.ReadFile("../e2e/run-e2e.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, function := range []string{
		"test_compatible_versions", "test_keda_compatible", "test_velero_compatible",
		"test_jaeger_unknown", "test_opentelemetry_unknown", "test_opentelemetry_operator_codiscovery",
	} {
		t.Run(function, func(t *testing.T) {
			body := regexp.MustCompile(`(?ms)^` + function + `\(\) \{\r?\n(.*?)^\}`).FindString(string(script))
			if body == "" {
				t.Fatal("missing scenario function")
			}
			if !strings.Contains(body, `assert_output_contains "^[[:space:]]*incompatibleTools: 0$"`) {
				t.Error("scenario must check the anchored YAML incompatibleTools summary")
			}
			if !strings.Contains(body, "assert_exit_code 0") {
				t.Error("scenario must also check exit status for upgrade incompatibilities")
			}
			if strings.Contains(body, `assert_output_not_contains "current_incompatibility"`) {
				t.Error("JSON-only field name cannot assert YAML incompatibility")
			}
		})
	}
}

func TestScanResultsSummaryMatchesE2EAssertion(t *testing.T) {
	zeroCurrent := regexp.MustCompile(`(?m)^[[:space:]]*incompatibleTools: 0$`)
	for _, tc := range []struct {
		name             string
		current, upgrade []Incompatibility
		wantZeroCurrent  bool
		wantUpgradeCount int
	}{
		{"no known incompatibilities", nil, nil, true, 0},
		{"current incompatibility despite zero upgrade", []Incompatibility{{Message: "current mismatch"}}, nil, false, 0},
		{"upgrade-only incompatibility", nil, []Incompatibility{{Message: "upgrade mismatch"}}, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, result, err := FormatScanResultsAsYAML([]DiscoveredTool{{
				Name: "fixture", CurrentIncompatibility: tc.current, UpgradeIncompatibility: tc.upgrade,
			}})
			if err != nil {
				t.Fatal(err)
			}
			if zeroCurrent.Match(data) != tc.wantZeroCurrent {
				t.Errorf("incorrect zero-current summary match:\n%s", data)
			}
			if result.Summary.UpgradeIncompatibleTools != tc.wantUpgradeCount {
				t.Errorf("upgrade count = %d, want %d", result.Summary.UpgradeIncompatibleTools, tc.wantUpgradeCount)
			}
		})
	}
}

func TestFluxKubernetesPatchMinimum(t *testing.T) {
	graph, err := LoadGraphFromYAML("compatibility/flux.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, flux := range []struct {
		version, belowMinimum, minimum, supportedPatch, excludedMinor string
	}{
		{"2.9.6", "1.34.0", "1.34.1", "1.34.9", "1.37.0"},
		{"2.8.0", "1.34.0", "1.34.1", "1.34.9", "1.36.0"},
		{"2.7.0", "1.34.0", "1.34.1", "1.34.9", "1.35.0"},
		{"2.2.0", "1.27.0", "1.27.1", "1.27.9", "1.29.0"},
		{"2.1.0", "1.27.0", "1.27.1", "1.27.9", "1.29.0"},
		{"2.0.0", "1.27.0", "1.27.1", "1.27.9", "1.28.0"},
	} {
		t.Run(flux.version, func(t *testing.T) {
			for _, tc := range []struct {
				version string
				want    int
			}{
				{flux.belowMinimum, NotCompatible},
				{flux.minimum, Compatible},
				{flux.supportedPatch, Compatible},
				{flux.excludedMinor, NotCompatible},
			} {
				t.Run(tc.version, func(t *testing.T) {
					version := semver.MustParse(tc.version)
					// Keep the data's patch minimum and the matcher's interpretation aligned.
					for _, edge := range graph.Edges {
						if edge.SourceVersion.Minor() == semver.MustParse(flux.version).Minor() && edge.DestinationVersionRange.Check(version) != (tc.want == Compatible) {
							t.Errorf("raw constraint %s: want compatibility %d", edge.DestinationVersionRange, tc.want)
						}
					}
					got, reason := graph.CheckCompatibility("flux", semver.MustParse(flux.version), Kubernetes, version)
					if got != tc.want {
						t.Errorf("got %d (%s), want %d", got, reason, tc.want)
					}
				})
			}
		})
	}
}
