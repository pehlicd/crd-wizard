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
package k8s

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/pehlicd/crd-wizard/internal/logger"
)

// ClusterManager manages multiple Kubernetes cluster connections.
// It loads all contexts from kubeconfig at startup and provides
// access to clients for each registered cluster.
type ClusterManager struct {
	clients        map[string]*Client // contextName -> Client
	contextNames   []string           // ordered list of context names
	currentContext string
	mu             sync.RWMutex
	log            *logger.Logger
}

// ClusterInfo represents basic information about a cluster for the API.
type ClusterEntry struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
}

// NewClusterManager creates a new ClusterManager and loads all contexts from kubeconfig.
// If kubeconfigPath is empty, it will use the default kubeconfig location.
// Invalid contexts are skipped with warnings rather than failing the entire initialization.
func NewClusterManager(kubeconfigPath string, log *logger.Logger) (*ClusterManager, error) {
	// Expand tilde in path
	if strings.HasPrefix(kubeconfigPath, "~/") {
		home := homedir.HomeDir()
		if home != "" {
			kubeconfigPath = filepath.Join(home, kubeconfigPath[2:])
		}
	}

	// Load kubeconfig to get all contexts
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		loadingRules.ExplicitPath = kubeconfigPath
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %w", err)
	}

	manager := &ClusterManager{
		clients:        make(map[string]*Client),
		contextNames:   make([]string, 0),
		currentContext: rawConfig.CurrentContext,
		log:            log,
	}

	// Load clients for all contexts
	for contextName := range rawConfig.Contexts {
		client, err := NewClient(kubeconfigPath, contextName, log)
		if err != nil {
			log.Warn("failed to load context, skipping", "context", contextName, "err", err)
			continue
		}
		manager.clients[contextName] = client
		manager.contextNames = append(manager.contextNames, contextName)
		log.Debug("loaded context", "context", contextName)
	}

	if len(manager.clients) == 0 {
		return nil, fmt.Errorf("no valid contexts found in kubeconfig")
	}

	// Validate that current context is loaded
	if _, ok := manager.clients[manager.currentContext]; !ok {
		// Fall back to first available context
		manager.currentContext = manager.contextNames[0]
		log.Warn("default context not available, using first available", "context", manager.currentContext)
	}

	log.Info("cluster manager initialized", "clusters", len(manager.clients), "current", manager.currentContext)

	return manager, nil
}

// GetClient returns the client for a specific cluster context.
func (m *ClusterManager) GetClient(name string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[name]
	if !ok {
		return nil, fmt.Errorf("cluster %q not found", name)
	}
	return client, nil
}

// GetCurrentClient returns the client for the current context.
func (m *ClusterManager) GetCurrentClient() *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.clients[m.currentContext]
}

// SetCurrentContext changes the current context.
func (m *ClusterManager) SetCurrentContext(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[name]; !ok {
		return fmt.Errorf("cluster %q not found", name)
	}
	m.currentContext = name
	return nil
}

// GetCurrentContextName returns the name of the current context.
func (m *ClusterManager) GetCurrentContextName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentContext
}

// ListClusters returns a list of all available clusters.
func (m *ClusterManager) ListClusters() []ClusterEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clusters := make([]ClusterEntry, 0, len(m.contextNames))
	for _, name := range m.contextNames {
		clusters = append(clusters, ClusterEntry{
			Name:      name,
			IsCurrent: name == m.currentContext,
		})
	}
	return clusters
}

// ClusterCount returns the number of loaded clusters.
func (m *ClusterManager) ClusterCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.clients)
}

// ContextNames returns a copy of the ordered list of context names.
func (m *ClusterManager) ContextNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, len(m.contextNames))
	copy(names, m.contextNames)
	return names
}
