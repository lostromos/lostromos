// Copyright 2017 the lostromos Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tmpl

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// CustomResource provides some helper methods for interacting with the
// kubernetes custom resource inside the templates.
type CustomResource struct {
	Resource *unstructured.Unstructured // represents the resource from kubernetes
}

// Name will return the Name from the custom resource
func (cr CustomResource) Name() string {
	return cr.Resource.GetName()
}

// GetField will traverse all the fields to return the string value of the
// requested field. If the field is not found it will return an empty string
func (cr CustomResource) GetField(fields ...string) string {
	if str, ok := getNestedField(cr.Resource.Object, fields...).(string); ok {
		return str
	}
	return ""
}

// copied from https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/unstructured/unstructured.go
func getNestedField(obj map[string]interface{}, fields ...string) interface{} {
	var val interface{} = obj
	for _, field := range fields {
		if _, ok := val.(map[string]interface{}); !ok {
			return nil
		}
		val = val.(map[string]interface{})[field]
	}
	return val
}
