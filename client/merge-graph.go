package client

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed compatibility/*.yaml
var compatibilityFS embed.FS

func merge2(dir string) (*Graph, error) {
	mergedGraph := Graph{}

	// Check if we're using embedded files or filesystem
	if dir == "embedded" {
		// Use embedded files
		err := fs.WalkDir(compatibilityFS, "compatibility", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Only process YAML files
			if !d.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
				data, readErr := compatibilityFS.ReadFile(path)
				if readErr != nil {
					return readErr
				}

				var graph Graph
				unmarshalErr := yaml.Unmarshal(data, &graph)
				if unmarshalErr != nil {
					return unmarshalErr
				}

				mergedGraph.Edges = append(mergedGraph.Edges, graph.Edges...)
				mergedGraph.Nodes = append(mergedGraph.Nodes, graph.Nodes...)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	} else {
		// Use filesystem
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only process YAML files
			if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
				graph, grapErr := LoadGraphFromYAML(path)
				if grapErr != nil {
					log.Fatalf("Failed to load compatibility graph: %v", grapErr)
				}
				mergedGraph.Edges = append(mergedGraph.Edges, graph.Edges...)
				mergedGraph.Nodes = append(mergedGraph.Nodes, graph.Nodes...)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return &mergedGraph, nil
}

// MergeYamlFiles merges all YAML files in a directory into one map
func MergeYamlFiles(dir string) (map[string][]interface{}, error) {
	mergedData := make(map[string][]interface{})

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process YAML files
		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
			fmt.Printf("Merging: %s\n", path)

			// Read YAML file
			fileData, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Unmarshal YAML into a temporary map
			var fileMap map[string][]interface{}
			err = yaml.Unmarshal(fileData, &fileMap)
			if err != nil {
				return err
			}

			// Merge fileMap into mergedData
			for key, value := range fileMap {
				initialData := mergedData[key]
				mergedData[key] = append(initialData, value)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mergedData, nil
}

// SaveMergedYaml writes the merged YAML data into a new file
func SaveMergedYaml(data *Graph, outputPath string) error {
	// Marshal the merged data into YAML
	mergedYaml, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// Write to file
	err = os.WriteFile(outputPath, mergedYaml, 0644)
	if err != nil {
		return err
	}

	return nil
}

func CreateMergeGraph() (*Graph, error) {
	// Try to use embedded files first
	mergedData, err := merge2("embedded")
	if err != nil {
		// Fall back to filesystem if embedded files fail
		fmt.Printf("Using filesystem compatibility files: %v\n", err)
		dir := "./compatibility"
		mergedData, err = merge2(dir)
		if err != nil {
			return nil, fmt.Errorf("error merging YAML files: %v", err)
		}
	} else {
		fmt.Println("Using embedded compatibility files")
	}

	return mergedData, nil
}
