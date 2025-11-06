/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package clustermanager

import (
	"fmt"
	"sync"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
)

// ClusterManager manages multiple Kubernetes cluster clients
type ClusterManager struct {
	clusters       map[string]*k8s.Client
	defaultCluster string
	mu             sync.RWMutex
	log            *logger.Logger
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(log *logger.Logger) *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*k8s.Client),
		log:      log,
	}
}

// AddCluster registers a new cluster client with the given name
func (cm *ClusterManager) AddCluster(name string, client *k8s.Client) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.clusters[name]; exists {
		return fmt.Errorf("cluster %s already exists", name)
	}

	cm.clusters[name] = client

	// Set as default if it's the first cluster
	if cm.defaultCluster == "" {
		cm.defaultCluster = name
	}

	cm.log.Info("cluster registered", "name", name)
	return nil
}

// GetClient retrieves a cluster client by name
func (cm *ClusterManager) GetClient(name string) (*k8s.Client, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clusters[name]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", name)
	}

	return client, nil
}

// GetDefaultClient returns the first registered client or nil if none exists
func (cm *ClusterManager) GetDefaultClient() *k8s.Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.defaultCluster != "" {
		if client, exists := cm.clusters[cm.defaultCluster]; exists {
			return client
		}
	}

	return nil
}

// ListClusters returns a list of all registered cluster names
func (cm *ClusterManager) ListClusters() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	names := make([]string, 0, len(cm.clusters))
	for name := range cm.clusters {
		names = append(names, name)
	}

	return names
}

// HasClusters returns true if at least one cluster is registered
func (cm *ClusterManager) HasClusters() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.clusters) > 0
}

// Count returns the number of registered clusters
func (cm *ClusterManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.clusters)
}
