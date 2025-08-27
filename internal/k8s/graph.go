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
	"context"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pehlicd/crd-wizard/internal/models"
)

type graphBuilder struct {
	client      *Client
	ctx         context.Context
	objectCache map[types.UID]unstructured.Unstructured
	ownerIndex  map[types.UID][]types.UID
	nodes       map[types.UID]models.Node
	edges       map[string]models.Edge
	queue       []types.UID
	visited     map[types.UID]bool
}

// GetResourceGraph builds and returns the relationship graph for a resource.
func (c *Client) GetResourceGraph(ctx context.Context, startUID string) (*models.ResourceGraph, error) {
	builder := &graphBuilder{
		client:      c,
		ctx:         ctx,
		objectCache: make(map[types.UID]unstructured.Unstructured),
		ownerIndex:  make(map[types.UID][]types.UID),
		nodes:       make(map[types.UID]models.Node),
		edges:       make(map[string]models.Edge),
		queue:       []types.UID{types.UID(startUID)},
		visited:     make(map[types.UID]bool),
	}

	if err := builder.buildCaches(); err != nil {
		return nil, fmt.Errorf("failed to build resource cache: %w", err)
	}

	if _, ok := builder.objectCache[types.UID(startUID)]; !ok {
		return nil, fmt.Errorf("resource with UID %s not found in cluster", startUID)
	}

	builder.traceGraph()

	return builder.getResourceGraph(), nil
}

// buildCaches scans the cluster for all resources and builds the object and owner caches.
func (b *graphBuilder) buildCaches() error {
	apiResourceLists, err := b.client.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		// The `ServerPreferredResources` endpoint can return partial results even on error.
		// We log the error but continue processing any resources that were returned.
		// This is often caused by aggregated API servers being unavailable.
		b.client.log.Warn("could not discover all server resources", "err", err)
	}

	var mu sync.Mutex
	g, ctx := errgroup.WithContext(b.ctx)
	g.SetLimit(10)

	for _, list := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, resource := range list.APIResources {
			// Filter out resources that cannot be listed or are sub-resources.
			if !isListable(resource.Verbs) || strings.Contains(resource.Name, "/") {
				continue
			}

			// Explicitly skip the deprecated 'componentstatuses' resource to avoid warnings.
			// This resource is not relevant for building an ownership graph.
			if gv.Group == "" && resource.Name == "componentstatuses" {
				continue
			}

			gvr := gv.WithResource(resource.Name)
			g.Go(func() error {
				objList, err := b.client.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
				if err != nil {
					// It's common to lack permissions for some resources (e.g., cluster-scoped ones),
					// so we log these as warnings and continue.
					b.client.log.Warn("could not list", "gvr", gvr, "err", err)
					return nil
				}

				mu.Lock()
				defer mu.Unlock()
				for _, item := range objList.Items {
					b.objectCache[item.GetUID()] = item
					for _, owner := range item.GetOwnerReferences() {
						b.ownerIndex[owner.UID] = append(b.ownerIndex[owner.UID], item.GetUID())
					}
				}
				return nil
			})
		}
	}
	return g.Wait()
}

// traceGraph performs a breadth-first search to build the graph.
func (b *graphBuilder) traceGraph() {
	for len(b.queue) > 0 {
		uid := b.queue[0]
		b.queue = b.queue[1:]

		if b.visited[uid] {
			continue
		}
		b.visited[uid] = true

		obj, ok := b.objectCache[uid]
		if !ok {
			continue
		}

		b.addNode(obj)

		// Trace parents (upwards)
		for _, owner := range obj.GetOwnerReferences() {
			b.addEdge(owner.UID, uid)
			b.queue = append(b.queue, owner.UID)
		}

		// Trace children (downwards) using the pre-built index
		if children, ok := b.ownerIndex[uid]; ok {
			for _, childUID := range children {
				b.addEdge(uid, childUID)
				b.queue = append(b.queue, childUID)
			}
		}
	}
}

func (b *graphBuilder) addNode(obj unstructured.Unstructured) {
	b.nodes[obj.GetUID()] = models.Node{
		ID:    string(obj.GetUID()),
		Label: obj.GetName(),
		Type:  obj.GetKind(),
	}
}

func (b *graphBuilder) addEdge(source, target types.UID) {
	edgeKey := fmt.Sprintf("%s->%s", source, target)
	b.edges[edgeKey] = models.Edge{
		Source: string(source),
		Target: string(target),
	}
}

func (b *graphBuilder) getResourceGraph() *models.ResourceGraph {
	graph := &models.ResourceGraph{
		Nodes: make([]models.Node, 0, len(b.nodes)),
		Edges: make([]models.Edge, 0, len(b.edges)),
	}
	for _, node := range b.nodes {
		graph.Nodes = append(graph.Nodes, node)
	}
	for _, edge := range b.edges {
		graph.Edges = append(graph.Edges, edge)
	}
	return graph
}

func isListable(verbs []string) bool {
	for _, verb := range verbs {
		if verb == "list" {
			return true
		}
	}
	return false
}
