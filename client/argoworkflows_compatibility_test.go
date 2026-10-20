package client

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestArgoWorkflowsContinuousTestedEnvelope(t *testing.T) {
	graph, err := LoadGraphFromYAML("compatibility/argoworkflows.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, release := range []struct {
		version  string
		min, max int
	}{
		{"4.1.5", 34, 36},
		{"4.0.0", 31, 33},
		{"3.7.0", 31, 33},
		{"3.6.0", 28, 31},
		{"3.5.0", 28, 30},
	} {
		for minor := release.min - 1; minor <= release.max+1; minor++ {
			for _, patch := range []int{0, 9} {
				version := fmt.Sprintf("1.%d.%d", minor, patch)
				t.Run(release.version+"/"+version, func(t *testing.T) {
					want := NotCompatible
					if minor >= release.min && minor <= release.max {
						want = Compatible
					}
					got, reason := graph.CheckCompatibility("argoworkflows", semver.MustParse(release.version), Kubernetes, semver.MustParse(version))
					if got != want {
						t.Errorf("got %d (%s), want %d", got, reason, want)
					}
				})
			}
		}
	}
}
