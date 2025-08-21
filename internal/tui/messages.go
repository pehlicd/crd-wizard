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
package tui

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pehlicd/crd-wizard/internal/models"
)

type showInstancesMsg struct{ crd models.CRD }
type showDetailsMsg struct {
	crd      models.CRD
	instance unstructured.Unstructured
}

type instancesLoadedMsg struct{ instances []unstructured.Unstructured }

type fullCRDLoadedMsg struct {
	def *apiextensionsv1.CustomResourceDefinition
}

type crdsLoadedMsg struct{ crds []models.CRD }

type goBackMsg struct{}
type errMsg struct{ err error }
