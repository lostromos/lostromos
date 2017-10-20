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

package printctlr

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Controller provides a crwatcher.ResourceController that prints the events out
// as the are received. It is a basic implementation that can be used for
// debuging. It also serves as an example for how you could implement your own
// crwatcher.ResourceController.
type Controller struct{}

// ResourceAdded will receive a custom resource when it is created and
// print that the CR was added
func (c Controller) ResourceAdded(r *unstructured.Unstructured) {
	fmt.Printf("CR added: %s\n", r.GetName())
}

// ResourceUpdated receives both an the old version and current version of a
// custom resource and will print out the the custom resource was changed
func (c Controller) ResourceUpdated(oldR, newR *unstructured.Unstructured) {
	fmt.Printf("CR changed: %s\n", newR.GetName())
}

// ResourceDeleted will receive a custom resource when it is deleted and
// print that the CR was deleted
func (c Controller) ResourceDeleted(r *unstructured.Unstructured) {
	fmt.Printf("CR deleted: %s\n", r.GetName())
}
