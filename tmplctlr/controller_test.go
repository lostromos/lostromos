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

package tmplctlr

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	basicResource = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "dory",
			},
			"spec": map[string]interface{}{
				"Name": "Dory",
				"From": "Finding Nemo",
				"By":   "Disney",
			},
		},
	}

	basicCR = &CustomResource{Resource: basicResource}

	templates = []templateFile{
		// T0.tmpl is a plain template file that just invokes T1.
		{"0_base.tmpl", `--- {{template "file1.tmpl" . }}`},
		// T1.tmpl defines a template, T1 that invokes T2.
		{"file1.tmpl", `name: {{ .GetField "metadata" "name"  }}-configmap`},
	}
)

// templateFile defines the contents of a template to be stored in a file, for testing.
type templateFile struct {
	name     string
	contents string
}

func createTestDir(files []templateFile) string {
	dir, err := ioutil.TempDir("", "template")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		f, err := os.Create(filepath.Join(dir, file.name))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err = io.WriteString(f, file.contents)
		if err != nil {
			log.Fatal(err)
		}
	}
	return dir
}

func ExampleController_ResourceAdded() {
	dir := createTestDir(templates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	c := &Controller{
		Client:       kubePrint{},
		templatePath: filepath.Join(dir, "*.tmpl"),
	}

	c.ResourceAdded(basicResource)
	// Output:
	// INFO: resource added, cr: dory
	// DEBUG: applied Kubernetes objects, cr: dory results: Kube Apply Called
}

func ExampleController_ResourceUpdated() {
	dir := createTestDir(templates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	c := &Controller{
		Client:       kubePrint{},
		templatePath: filepath.Join(dir, "*.tmpl"),
	}

	c.ResourceUpdated(basicResource, basicResource)
	// Output:
	// INFO: resource updated, cr: dory
	// DEBUG: applied Kubernetes objects, cr: dory results: Kube Apply Called
}

func ExampleController_ResourceDeleted() {
	dir := createTestDir(templates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	c := &Controller{
		Client:       kubePrint{},
		templatePath: filepath.Join(dir, "*.tmpl"),
	}

	c.ResourceDeleted(basicResource)
	// Output:
	// INFO: resource deleted, cr: dory
	// DEBUG: deleted Kubernetes objects, cr: dory results: Kube Delete Called
}
