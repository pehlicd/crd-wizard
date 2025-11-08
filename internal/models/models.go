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
package models

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// APICRD is used for the web API, to include the full spec.
type APICRD struct {
	APIVersion    string                                       `json:"apiVersion"`
	Kind          string                                       `json:"kind"`
	Metadata      metav1.ObjectMeta                            `json:"metadata"`
	Spec          apiextensionsv1.CustomResourceDefinitionSpec `json:"spec"`
	InstanceCount int                                          `json:"instanceCount"`
}

// CRD model is used for the TUI, which only needs a subset of fields.
// Despite the name, this can represent any Kubernetes resource (both CRD-based and built-in).
type CRD struct {
	APIVersion    string `json:"apiVersion"`
	Kind          string `json:"kind"`
	Name          string `json:"name"` // For CRDs: the CRD name. For others: plural resource name
	Group         string `json:"group"`
	Version       string `json:"version"`  // API version
	Resource      string `json:"resource"` // Plural resource name (e.g., "pods", "deployments")
	Scope         string `json:"scope"`    // "Namespaced" or "Cluster"
	InstanceCount int    `json:"instanceCount"`
	Namespaced    bool   `json:"namespaced"` // Whether resource is namespaced
}

func FromK8sCRD(k8sCrd apiextensionsv1.CustomResourceDefinition, instanceCount int) CRD {
	// Determine the storage version
	version := ""
	for _, v := range k8sCrd.Spec.Versions {
		if v.Storage {
			version = v.Name
			break
		}
	}
	if version == "" && len(k8sCrd.Spec.Versions) > 0 {
		version = k8sCrd.Spec.Versions[0].Name
	}

	namespaced := k8sCrd.Spec.Scope == apiextensionsv1.NamespaceScoped

	return CRD{
		APIVersion:    k8sCrd.APIVersion,
		Kind:          k8sCrd.Spec.Names.Kind,
		Name:          k8sCrd.Name,
		Group:         k8sCrd.Spec.Group,
		Version:       version,
		Resource:      k8sCrd.Spec.Names.Plural,
		Scope:         string(k8sCrd.Spec.Scope),
		InstanceCount: instanceCount,
		Namespaced:    namespaced,
	}
}

// FromAPIResource creates a CRD model from a discovered API resource.
// This enables the tool to work with all Kubernetes resources, not just CRDs.
func FromAPIResource(gv schema.GroupVersion, resource metav1.APIResource, instanceCount int) CRD {
	scope := "Cluster"
	if resource.Namespaced {
		scope = "Namespaced"
	}

	// For built-in resources (empty group), use the version as part of APIVersion
	apiVersion := gv.String()
	if gv.Group == "" {
		apiVersion = gv.Version
	}

	return CRD{
		APIVersion:    apiVersion,
		Kind:          resource.Kind,
		Name:          resource.Name, // Use plural name as identifier
		Group:         gv.Group,
		Version:       gv.Version,
		Resource:      resource.Name,
		Scope:         scope,
		InstanceCount: instanceCount,
		Namespaced:    resource.Namespaced,
	}
}

// GVR returns the GroupVersionResource for this resource.
func (c CRD) GVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    c.Group,
		Version:  c.Version,
		Resource: c.Resource,
	}
}

func ToAPICRD(k8sCrd apiextensionsv1.CustomResourceDefinition, instanceCount int) APICRD {
	metadata := k8sCrd.ObjectMeta
	metadata.ManagedFields = nil
	return APICRD{
		APIVersion:    k8sCrd.APIVersion,
		Kind:          k8sCrd.Kind,
		Metadata:      metadata,
		Spec:          k8sCrd.Spec,
		InstanceCount: instanceCount,
	}
}

// ResourceGraph represents the structure for the graph API response.
type ResourceGraph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a single Kubernetes resource in the graph.
type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// Edge represents a relationship between two nodes in the graph.
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// ClusterInfo holds information about the Kubernetes cluster.
type ClusterInfo struct {
	ClusterName   string `json:"clusterName"`
	ServerVersion string `json:"serverVersion"`
	NumCRDs       int    `json:"numCRDs"` // For backward compatibility, renamed to NumResources internally
	NumResources  int    `json:"numResources"`
}
