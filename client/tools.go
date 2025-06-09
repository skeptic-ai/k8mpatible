package client

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v2"
)

var (
	Logger = getSlogger()
)

func getSlogger() *slog.Logger {
	options := &slog.HandlerOptions{AddSource: true}
	handler := slog.NewTextHandler(os.Stdout, options)

	return slog.New(handler)
}

// LoadGraphFromYAML loads the compatibility graph from a YAML file.
func LoadGraphFromYAML(filename string) (*Graph, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var graph Graph
	err = yaml.Unmarshal(data, &graph)
	if err != nil {
		return nil, err
	}

	return &graph, nil
}

// CheckCompatibility checks if two tools with specific versions are compatible.
func (g *Graph) CheckCompatibility(sourceName string, sourceVersion *semver.Version, destName string, destVersion *semver.Version) (int, string) {
	for _, edge := range g.Edges {

		if edge.SourceName == sourceName && edge.DestinationName == destName {
			// increment both by major version +1 so we can use equal to comapre and eliminate the need for a range
			edgeIncremented := edge.SourceVersion.IncMinor()
			destIncremented := sourceVersion.IncMinor()
			if edgeIncremented.Equal(&destIncremented) {
				Logger.Info("Sources matched for", "range", edge.DestinationVersionRange, "version", destVersion)
				destVersionNormalized := semver.New(destVersion.Major(), destVersion.Minor(), 0, "", "")
				Logger.Info("Normalized version", "version", destVersionNormalized)
				matched, err := edge.DestinationVersionRange.Validate(destVersionNormalized)

				if len(err) > 0 {
					// fmt.Println("error validating version range: %s", err)
					return NotCompatible, fmt.Sprintf("requires range %s", edge.DestinationVersionRange)
				}

				if matched {
					if edge.Compatible {
						return Compatible, edge.Reason
					} else {
						return NotCompatible, edge.Reason
					}
				} else {
					Logger.Info("no ranges matched")
					return NotCompatible, "No edges matched"
				}

			}
		}
	}
	return Unknown, "no edges found"
}

func GetActiveVersion(tools []DiscoveredTool, name string) *semver.Version {
	for _, tool := range tools {
		if tool.Name == name {
			return tool.Version
		}
	}
	return nil
}

const (
	Compatible = iota
	NotCompatible
	Unknown
)
