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

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func ExampleController_ResourceAdded() {
	r := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}

	Controller{}.ResourceAdded(r)
	// Output:
	// CR added: Thing1
}

func ExampleController_ResourceUpdated() {
	oldR := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}
	newR := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing2",
			},
		},
	}

	Controller{}.ResourceUpdated(oldR, newR)
	// Output:
	// CR changed: Thing2
}

func ExampleController_ResourceDeleted() {
	r := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}

	Controller{}.ResourceDeleted(r)
	// Output:
	// CR deleted: Thing1
}
