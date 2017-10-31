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
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/wpengine/lostromos/metrics"
	"github.com/wpengine/lostromos/tmpl"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Controller implements a valid crwatcher.ResourceController that will manage
// resources in kubernetes based on the provided template files.
type Controller struct {
	templatePath string     //path to dir where templates are located
	Client       KubeClient //client for talking with kubernetes
}

// NewController will return a configured Controller
func NewController(tmplDir string, kubeCfg string) *Controller {
	c := &Controller{
		Client:       &Kubectl{ConfigFile: kubeCfg},
		templatePath: filepath.Join(tmplDir, "*.tmpl"),
	}
	return c
}

// ResourceAdded is called when a custom resource is created and will generate
// the template files and apply them to Kubernetes
func (c Controller) ResourceAdded(r *unstructured.Unstructured) {
	fmt.Printf("INFO: resource added, cr: %s\n", r.GetName())
	c.apply(r)
	metrics.CreatedReleases.Inc()
	metrics.ManagedReleases.Inc()
	metrics.TotalEvents.Inc()
}

// ResourceUpdated is called when a custom resource is updated or during a
// resync and will generate the template files and apply them to Kubernetes
func (c Controller) ResourceUpdated(oldR, newR *unstructured.Unstructured) {
	fmt.Printf("INFO: resource updated, cr: %s\n", newR.GetName())
	c.apply(newR)
	metrics.UpdatedReleases.Inc()
	metrics.TotalEvents.Inc()
}

func (c Controller) apply(r *unstructured.Unstructured) {
	cr := &tmpl.CustomResource{
		Resource: r,
	}
	tmpFile, err := ioutil.TempFile("", "lostromos")
	if err != nil {
		fmt.Printf("ERROR: failed to get tmp file error: %s\n", err)
		return
	}
	err = tmpl.Parse(cr, c.templatePath, tmpFile)
	if err != nil {
		fmt.Printf("ERROR: failed to generate template error: %s\n", err)
		return
	}
	out, err := c.Client.Apply(tmpFile.Name())
	if err != nil {
		fmt.Printf("ERROR: failed to apply template error: %s - [ %s ]\n", err, out)
		fmt.Printf("DEBUG: template to apply: \n%s", readFile(tmpFile.Name()))
		return
	}
	fmt.Printf("DEBUG: applied Kubernetes objects, cr: %s results: %s\n", r.GetName(), out)
}

// ResourceDeleted is called when a custom resource is created and will generate
// the template files and delete them from Kubernetes
func (c Controller) ResourceDeleted(r *unstructured.Unstructured) {
	cr := &tmpl.CustomResource{
		Resource: r,
	}
	fmt.Printf("INFO: resource deleted, cr: %s\n", r.GetName())
	tmpFile, err := ioutil.TempFile("", "lostromos")
	if err != nil {
		fmt.Printf("ERROR: failed to get tmp file error: %s\n", err)
		return
	}
	err = tmpl.Parse(cr, c.templatePath, tmpFile)
	if err != nil {
		fmt.Printf("ERROR: failed to generate template error: %s\n", err)
		return
	}
	out, err := c.Client.Delete(tmpFile.Name())
	if err != nil {
		fmt.Printf("ERROR: failed to delete template error: %s - [ %s ]\n", err, out)
		fmt.Printf("DEBUG: template to delete: \n%s", readFile(tmpFile.Name()))
		return
	}
	fmt.Printf("DEBUG: deleted Kubernetes objects, cr: %s results: %s\n", r.GetName(), out)
	metrics.DeletedReleases.Inc()
	metrics.ManagedReleases.Dec()
	metrics.TotalEvents.Inc()
}

func readFile(filepath string) string {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	return string(content[:])
}
