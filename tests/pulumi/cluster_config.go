package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ClusterConfig holds configuration for each cluster
type ClusterConfig struct {
	Name       string
	KindConfig string
}

// getClusterConfig dynamically discovers cluster configuration
// by reading from flux/clusters/<stack>/kind.yaml
func getClusterConfig(stack string) (*ClusterConfig, error) {
	// Base path to cluster configurations
	clustersDir := "../../flux/clusters"
	clusterPath := filepath.Join(clustersDir, stack)
	kindConfigPath := filepath.Join(clusterPath, "kind.yaml")

	// Check if cluster directory exists
	if _, err := os.Stat(clusterPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cluster '%s' not found in %s", stack, clustersDir)
	}

	// Check if kind.yaml exists
	if _, err := os.Stat(kindConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kind.yaml not found for cluster '%s' at %s", stack, kindConfigPath)
	}

	return &ClusterConfig{
		Name:       stack,
		KindConfig: kindConfigPath,
	}, nil
}
