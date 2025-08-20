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
	"sync"
	"time"

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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/pehlicd/crd-explorer/internal/models"
)

type Client struct {
	ExtensionsClient *apiextensionsclientset.Clientset
	DynamicClient    dynamic.Interface
	CoreClient       *kubernetes.Clientset
	DiscoveryClient  discovery.DiscoveryInterface
	ApiExtClient     *apiextensionsclientset.Clientset
}

func NewClient(kubeconfigPath, contextName string) (*Client, error) {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		if strings.HasPrefix(kubeconfigPath, "~/") {
			home := homedir.HomeDir()
			if home == "" {
				return nil, fmt.Errorf("cannot expand tilde path: user home directory not found")
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
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("error building config for context %q from path %q: %w", contextName, kubeconfigPath, err)
		}
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
		ApiExtClient:     apiExtClient,
	}, nil
}

func (c *Client) GetCRDs(ctx context.Context) ([]models.CRD, error) {
	crdList, err := c.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CRDs: %w", err)
	}
	uiCrds := make([]models.CRD, len(crdList.Items))
	var wg sync.WaitGroup
	for i, crd := range crdList.Items {
		wg.Add(1)
		go func(i int, crd apiextensionsv1.CustomResourceDefinition) {
			defer wg.Done()
			instanceCount := c.CountCRDInstances(ctx, crd)
			uiCrds[i] = models.FromK8sCRD(crd, instanceCount)
		}(i, crd)
	}
	wg.Wait()
	return uiCrds, nil
}

func (c *Client) GetCRDsByKind(ctx context.Context, kind string) ([]models.CRD, error) {
	allCRDsList, err := c.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CRDs: %w", err)
	}

	var matchingCRDs []apiextensionsv1.CustomResourceDefinition
	for _, crd := range allCRDsList.Items {
		if crd.Spec.Names.Kind == kind {
			matchingCRDs = append(matchingCRDs, crd)
		}
	}

	if len(matchingCRDs) == 0 {
		return []models.CRD{}, nil
	}

	filteredCRDs := make([]models.CRD, len(matchingCRDs))
	var wg sync.WaitGroup
	for i, crd := range matchingCRDs {
		wg.Add(1)
		go func(i int, crd apiextensionsv1.CustomResourceDefinition) {
			defer wg.Done()
			instanceCount := c.CountCRDInstances(ctx, crd)
			filteredCRDs[i] = models.FromK8sCRD(crd, instanceCount)
		}(i, crd)
	}
	wg.Wait()

	return filteredCRDs, nil
}

func (c *Client) GetCRDsByName(ctx context.Context, name string) (*models.CRD, error) {
	filteredCRD, err := c.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CRDs with filter: %w", err)
	}

	crd := models.FromK8sCRD(*filteredCRD, c.CountCRDInstances(ctx, *filteredCRD))

	return &crd, nil
}

func (c *Client) GetCRsForCRD(ctx context.Context, crdName string) ([]unstructured.Unstructured, error) {
	crd, err := c.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CRD %s: %w", crdName, err)
	}
	gvr, _ := getGVRFromCRD(*crd)
	if gvr.Resource == "" {
		return nil, fmt.Errorf("could not determine GVR for CRD %s", crdName)
	}
	list, err := c.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list instances for CRD %s: %w", crdName, err)
	}
	return list.Items, nil
}

func (c *Client) GetSingleCR(ctx context.Context, crdName, namespace, name string) (*unstructured.Unstructured, error) {
	crd, err := c.ExtensionsClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CRD %s: %w", crdName, err)
	}
	gvr, _ := getGVRFromCRD(*crd)
	if gvr.Resource == "" {
		return nil, fmt.Errorf("could not determine GVR for CRD %s", crdName)
	}
	var resource dynamic.ResourceInterface
	if crd.Spec.Scope == apiextensionsv1.NamespaceScoped {
		//if namespace == "_cluster" {
		//	return nil, fmt.Errorf("cannot use '_cluster' for namespaced resource")
		//}
		resource = c.DynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resource = c.DynamicClient.Resource(gvr)
	}
	return resource.Get(ctx, name, metav1.GetOptions{})
}

// GetFullCRD retrieves the complete CustomResourceDefinition object from the cluster.
func (c *Client) GetFullCRD(ctx context.Context, name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	crd, err := c.ApiExtClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
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
	list, err := c.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{TimeoutSeconds: &[]int64{5}[0]})
	if err != nil {
		return 0
	}
	return len(list.Items)
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

func HumanReadableAge(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	d := time.Since(t)
	if d.Hours() > 24*365 {
		return fmt.Sprintf("%.0fy", d.Hours()/(24*365))
	}
	if d.Hours() > 24*30 {
		return fmt.Sprintf("%.0fmo", d.Hours()/(24*30))
	}
	if d.Hours() > 24 {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}
