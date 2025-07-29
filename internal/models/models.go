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
type CRD struct {
	APIVersion    string `json:"apiVersion"`
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Group         string `json:"group"`
	Scope         string `json:"scope"`
	InstanceCount int    `json:"instanceCount"`
}

func FromK8sCRD(k8sCrd apiextensionsv1.CustomResourceDefinition, instanceCount int) CRD {
	return CRD{
		APIVersion:    k8sCrd.APIVersion,
		Kind:          k8sCrd.Spec.Names.Kind,
		Name:          k8sCrd.Name,
		Group:         k8sCrd.Spec.Group,
		Scope:         string(k8sCrd.Spec.Scope),
		InstanceCount: instanceCount,
	}
}

func ToAPICRD(k8sCrd apiextensionsv1.CustomResourceDefinition, instanceCount int) APICRD {
	return APICRD{
		APIVersion:    k8sCrd.APIVersion,
		Kind:          k8sCrd.Kind,
		Metadata:      k8sCrd.ObjectMeta,
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
