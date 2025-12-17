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
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/pehlicd/crd-wizard/internal/logger"
	"github.com/pehlicd/crd-wizard/internal/models"
)

type Client struct {
	ExtensionsClient *apiextensionsclientset.Clientset
	DynamicClient    dynamic.Interface
	CoreClient       *kubernetes.Clientset
	DiscoveryClient  discovery.DiscoveryInterface
	APIExtClient     *apiextensionsclientset.Clientset
	ClusterName      string
	log              *logger.Logger
}

func NewClient(kubeconfigPath, contextName string, log *logger.Logger) (*Client, error) {
	config, clusterName, err := buildConfig(kubeconfigPath, contextName)
	if err != nil {
		log.Error("error building config", "err", err)
		return nil, err
	}

	config.QPS = 100
	config.Burst = 150

	extensionsClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating extensions clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating core clientset: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating discovery client: %w", err)
	}

	apiExtClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	return &Client{
		ExtensionsClient: extensionsClient,
		DynamicClient:    dynamicClient,
		CoreClient:       coreClient,
		DiscoveryClient:  discoveryClient,
		APIExtClient:     apiExtClient,
		ClusterName:      clusterName,
		log:              log,
	}, nil
}

func buildConfig(kubeconfigPath, contextName string) (*rest.Config, string, error) {
	// First, try in-cluster config
	config, err := rest.InClusterConfig()
	if err == nil {
		// For in-cluster, there's no kubeconfig context, so we return a default name
		return config, "in-cluster", nil
	}

	// Fallback to out-of-cluster config
	if strings.HasPrefix(kubeconfigPath, "~/") {
		home := homedir.HomeDir()
		if home == "" {
			return nil, "", fmt.Errorf("cannot expand tilde path: user home directory not found")
		}
		kubeconfigPath = filepath.Join(home, kubeconfigPath[2:])
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		loadingRules.ExplicitPath = kubeconfigPath
	}

	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	// Get the client config
	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("error building client config for context %q from path %q: %w", contextName, kubeconfigPath, err)
	}

	// Get the raw config to find the cluster name from the context
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, "", fmt.Errorf("error getting raw kubeconfig: %w", err)
	}

	// Determine which context is being used
	currentContext := rawConfig.CurrentContext
	if contextName != "" {
		currentContext = contextName
	}

	c, ok := rawConfig.Contexts[currentContext]
	if !ok {
		return nil, "", fmt.Errorf("context %q not found in kubeconfig", currentContext)
	}

	clusterName := c.Cluster

	return clientConfig, clusterName, nil
}

func (c *Client) GetClusterInfo() (models.ClusterInfo, error) {
	versionInfo, err := c.DiscoveryClient.ServerVersion()
	if err != nil {
		return models.ClusterInfo{}, fmt.Errorf("failed to get server version: %w", err)
	}

	// Count all API resources using Discovery API
	apiResourceLists, err := c.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		c.log.Warn("could not discover all server resources", "err", err)
	}

	resourceCount := 0
	for _, list := range apiResourceLists {
		for _, resource := range list.APIResources {
			// Skip subresources
			if !strings.Contains(resource.Name, "/") {
				resourceCount++
			}
		}
	}

	return models.ClusterInfo{
		ClusterName:   c.ClusterName,
		ServerVersion: versionInfo.GitVersion,
		NumCRDs:       resourceCount, // Kept for backward compatibility
		NumResources:  resourceCount,
	}, nil
}

func (c *Client) GetCRDs(ctx context.Context) ([]models.CRD, error) {
	// Use Discovery API to get all API resources in the cluster
	apiResourceLists, err := c.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		// ServerPreferredResources can return partial results even on error
		// This is often caused by aggregated API servers being unavailable
		c.log.Warn("could not discover all server resources, using partial results", "err", err)
	}

	if apiResourceLists == nil {
		return nil, fmt.Errorf("failed to discover any API resources")
	}

	// Collect all resources that can be listed
	var resources []struct {
		gv       schema.GroupVersion
		resource metav1.APIResource
	}

	for _, list := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			c.log.Warn("could not parse group version", "groupVersion", list.GroupVersion, "err", err)
			continue
		}

		for _, resource := range list.APIResources {
			// Skip subresources (e.g., pods/log, deployments/status)
			if strings.Contains(resource.Name, "/") {
				continue
			}

			// Skip resources that can't be listed
			if !contains(resource.Verbs, "list") {
				continue
			}

			// Skip componentstatuses (deprecated and not useful)
			if gv.Group == "" && resource.Name == "componentstatuses" {
				continue
			}

			resources = append(resources, struct {
				gv       schema.GroupVersion
				resource metav1.APIResource
			}{gv: gv, resource: resource})
		}
	}

	// Create CRD models with instance counts
	uiCrds := make([]models.CRD, len(resources))
	var g errgroup.Group
	g.SetLimit(20) // Limit concurrent goroutines

	for i, res := range resources {
		i, res := i, res
		g.Go(func() error {
			gvr := schema.GroupVersionResource{
				Group:    res.gv.Group,
				Version:  res.gv.Version,
				Resource: res.resource.Name,
			}
			instanceCount := c.countResourceInstances(ctx, gvr)
			uiCrds[i] = models.FromAPIResource(res.gv, res.resource, instanceCount)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return uiCrds, nil
}

func (c *Client) GetCRsForCRD(ctx context.Context, crdName string) ([]unstructured.Unstructured, error) {
	// crdName is now the resource name (plural), e.g., "pods", "deployments", "certificates"
	// We need to find the GVR for this resource using Discovery API
	gvr, err := c.findGVRForResource(ctx, crdName)
	if err != nil {
		return nil, fmt.Errorf("failed to find resource %s: %w", crdName, err)
	}

	list, err := c.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list instances for resource %s: %w", crdName, err)
	}
	return list.Items, nil
}

func (c *Client) GetSingleCR(ctx context.Context, crdName, namespace, name string) (*unstructured.Unstructured, error) {
	// crdName is now the resource name (plural)
	gvr, namespaced, err := c.findGVRAndScopeForResource(ctx, crdName)
	if err != nil {
		return nil, fmt.Errorf("failed to find resource %s: %w", crdName, err)
	}

	var resource dynamic.ResourceInterface
	if namespaced {
		resource = c.DynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = c.DynamicClient.Resource(gvr)
	}

	unstructuredCR, err := resource.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	unstructured.RemoveNestedField(unstructuredCR.Object, "metadata", "managedFields")

	return unstructuredCR, nil
}

// GetFullCRD retrieves the complete CustomResourceDefinition object from the cluster.
func (c *Client) GetFullCRD(ctx context.Context, name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd, err := c.APIExtClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return crd, nil
}

func (c *Client) GetEvents(ctx context.Context, crdName, resourceUID string) ([]corev1.Event, error) {
	if resourceUID != "" {
		return c.getEventsForUID(ctx, resourceUID)
	}
	if crdName != "" {
		return c.getEventsForCRD(ctx, crdName)
	}
	return nil, fmt.Errorf("either crdName or resourceUid query parameter is required")
}

func (c *Client) getEventsForUID(ctx context.Context, uid string) ([]corev1.Event, error) {
	allEvents, err := c.CoreClient.CoreV1().Events("").List(ctx, metav1.ListOptions{TimeoutSeconds: &[]int64{10}[0]})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	var relatedEvents []corev1.Event
	targetUID := types.UID(uid)
	for _, event := range allEvents.Items {
		if event.InvolvedObject.UID == targetUID {
			relatedEvents = append(relatedEvents, event)
		}
	}
	return relatedEvents, nil
}

func (c *Client) getEventsForCRD(ctx context.Context, crdName string) ([]corev1.Event, error) {
	crList, err := c.GetCRsForCRD(ctx, crdName)
	if err != nil {
		return nil, err
	}
	if len(crList) == 0 {
		return []corev1.Event{}, nil
	}
	crUIDs := make(map[types.UID]bool)
	for _, item := range crList {
		crUIDs[item.GetUID()] = true
	}
	allEvents, err := c.CoreClient.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	var relatedEvents []corev1.Event
	for _, event := range allEvents.Items {
		if crUIDs[event.InvolvedObject.UID] {
			relatedEvents = append(relatedEvents, event)
		}
	}
	return relatedEvents, nil
}

func (c *Client) CountCRDInstances(ctx context.Context, crd apiextensionsv1.CustomResourceDefinition) int {
	gvr, _ := getGVRFromCRD(crd)
	if gvr.Resource == "" {
		return 0
	}
	return c.countResourceInstances(ctx, gvr)
}

func (c *Client) countResourceInstances(ctx context.Context, gvr schema.GroupVersionResource) int {
	list, err := c.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{TimeoutSeconds: &[]int64{5}[0]})
	if err != nil {
		return 0
	}
	return len(list.Items)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// findGVRForResource finds the GroupVersionResource for a given resource name using Discovery API.
// It returns the preferred version of the resource.
func (c *Client) findGVRForResource(ctx context.Context, resourceName string) (schema.GroupVersionResource, error) {
	gvr, _, err := c.findGVRAndScopeForResource(ctx, resourceName)
	return gvr, err
}

// findGVRAndScopeForResource finds the GroupVersionResource and scope for a given resource name.
func (c *Client) findGVRAndScopeForResource(ctx context.Context, resourceName string) (schema.GroupVersionResource, bool, error) {
	apiResourceLists, err := c.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		c.log.Warn("could not discover all server resources", "err", err)
	}

	if apiResourceLists == nil {
		return schema.GroupVersionResource{}, false, fmt.Errorf("no API resources discovered")
	}

	// Search for the resource in all API groups
	// Prefer the most recent version if multiple exist
	var foundGVR schema.GroupVersionResource
	var foundNamespaced bool
	var found bool

	for _, list := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range list.APIResources {
			if resource.Name == resourceName {
				foundGVR = schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}
				foundNamespaced = resource.Namespaced
				found = true
				// We found a match; use the first one we encounter
				// (ServerPreferredResources returns preferred versions first)
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return schema.GroupVersionResource{}, false, fmt.Errorf("resource %s not found in cluster", resourceName)
	}

	return foundGVR, foundNamespaced, nil
}

func getGVRFromCRD(crd apiextensionsv1.CustomResourceDefinition) (schema.GroupVersionResource, string) {
	storageVersion := ""
	for _, v := range crd.Spec.Versions {
		if v.Storage {
			storageVersion = v.Name
			break
		}
	}
	if storageVersion == "" && len(crd.Spec.Versions) > 0 {
		storageVersion = crd.Spec.Versions[0].Name
	}
	if storageVersion != "" {
		return schema.GroupVersionResource{Group: crd.Spec.Group, Version: storageVersion, Resource: crd.Spec.Names.Plural}, storageVersion
	}
	return schema.GroupVersionResource{}, ""
}
