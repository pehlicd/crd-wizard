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
	"os"
	"testing"

	"github.com/pehlicd/crd-wizard/internal/k8s"
	"github.com/pehlicd/crd-wizard/internal/logger"
)

func TestNewClusterManager(t *testing.T) {
	log := logger.NewLogger("text", "info", os.Stderr)
	cm := NewClusterManager(log)

	if cm == nil {
		t.Fatal("NewClusterManager returned nil")
	}

	if cm.Count() != 0 {
		t.Errorf("Expected 0 clusters, got %d", cm.Count())
	}

	if cm.HasClusters() {
		t.Error("Expected HasClusters to return false")
	}
}

func TestAddCluster(t *testing.T) {
	log := logger.NewLogger("text", "info", os.Stderr)
	cm := NewClusterManager(log)

	client := &k8s.Client{ClusterName: "test-cluster"}

	err := cm.AddCluster("test-cluster", client)
	if err != nil {
		t.Fatalf("AddCluster failed: %v", err)
	}

	if cm.Count() != 1 {
		t.Errorf("Expected 1 cluster, got %d", cm.Count())
	}

	if !cm.HasClusters() {
		t.Error("Expected HasClusters to return true")
	}
}

func TestAddClusterDuplicate(t *testing.T) {
	log := logger.NewLogger("text", "info", os.Stderr)
	cm := NewClusterManager(log)

	client := &k8s.Client{ClusterName: "test-cluster"}

	err := cm.AddCluster("test-cluster", client)
	if err != nil {
		t.Fatalf("First AddCluster failed: %v", err)
	}

	err = cm.AddCluster("test-cluster", client)
	if err == nil {
		t.Error("Expected error when adding duplicate cluster")
	}
}

func TestGetClient(t *testing.T) {
	log := logger.NewLogger("text", "info", os.Stderr)
	cm := NewClusterManager(log)

	client := &k8s.Client{ClusterName: "test-cluster"}
	err := cm.AddCluster("test-cluster", client)
	if err != nil {
		t.Fatalf("AddCluster failed: %v", err)
	}

	retrievedClient, err := cm.GetClient("test-cluster")
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}

	if retrievedClient != client {
		t.Error("Retrieved client does not match added client")
	}
}

func TestGetClientNotFound(t *testing.T) {
	log := logger.NewLogger("text", "info", os.Stderr)
	cm := NewClusterManager(log)

	_, err := cm.GetClient("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent cluster")
	}
}

func TestGetDefaultClient(t *testing.T) {
	log := logger.NewLogger("text", "info", os.Stderr)
	cm := NewClusterManager(log)

	defaultClient := cm.GetDefaultClient()
	if defaultClient != nil {
		t.Error("Expected nil when no clusters are registered")
	}

	client := &k8s.Client{ClusterName: "test-cluster"}
	err := cm.AddCluster("test-cluster", client)
	if err != nil {
		t.Fatalf("AddCluster failed: %v", err)
	}

	defaultClient = cm.GetDefaultClient()
	if defaultClient == nil {
		t.Error("Expected non-nil default client")
	}
}

func TestListClusters(t *testing.T) {
	log := logger.NewLogger("text", "info", os.Stderr)
	cm := NewClusterManager(log)

	clusters := cm.ListClusters()
	if len(clusters) != 0 {
		t.Errorf("Expected 0 clusters, got %d", len(clusters))
	}

	client1 := &k8s.Client{ClusterName: "cluster1"}
	client2 := &k8s.Client{ClusterName: "cluster2"}

	cm.AddCluster("cluster1", client1)
	cm.AddCluster("cluster2", client2)

	clusters = cm.ListClusters()
	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}

	found1, found2 := false, false
	for _, name := range clusters {
		if name == "cluster1" {
			found1 = true
		}
		if name == "cluster2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Not all clusters found in ListClusters result")
	}
}
