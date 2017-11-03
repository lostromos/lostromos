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
	"io/ioutil"
	"path/filepath"

	"github.com/wpengine/lostromos/metrics"
	"github.com/wpengine/lostromos/tmpl"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Controller implements a valid crwatcher.ResourceController that will manage
// resources in kubernetes based on the provided template files.
type Controller struct {
	templatePath string     //path to dir where templates are located
	Client       KubeClient //client for talking with kubernetes
	logger       *zap.SugaredLogger
}

// NewController will return a configured Controller
func NewController(tmplDir string, kubeCfg string, logger *zap.SugaredLogger) *Controller {
	if logger == nil {
		// If you don't give us a logger, set logger to a nop logger
		logger = zap.NewNop().Sugar()
	}
	c := &Controller{
		Client:       &Kubectl{ConfigFile: kubeCfg},
		templatePath: filepath.Join(tmplDir, "*.tmpl"),
		logger:       logger,
	}
	return c
}

// ResourceAdded is called when a custom resource is created and will generate
// the template files and apply them to Kubernetes
func (c Controller) ResourceAdded(r *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource added", "resource", r.GetName())
	c.apply(r)
	metrics.CreatedReleases.Inc()
	metrics.ManagedReleases.Inc()
}

// ResourceUpdated is called when a custom resource is updated or during a
// resync and will generate the template files and apply them to Kubernetes
func (c Controller) ResourceUpdated(oldR, newR *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource updated", "resource", newR.GetName())
	c.apply(newR)
	metrics.UpdatedReleases.Inc()
}

func (c Controller) apply(r *unstructured.Unstructured) {
	cr := &tmpl.CustomResource{
		Resource: r,
	}
	tmpFile, err := ioutil.TempFile("", "lostromos")
	if err != nil {
		c.logger.Errorw("failed to get tmp file", "error", err)
		return
	}
	err = tmpl.Parse(cr, c.templatePath, tmpFile)
	if err != nil {
		c.logger.Errorw("failed to generate template error", "error", err)
		return
	}
	out, err := c.Client.Apply(tmpFile.Name())
	if err != nil {
		c.logger.Errorw("failed to apply template", "error", err, "result", out)
		c.logger.Debugw("template we want to apply", "template", readFile(tmpFile.Name()), "fileName", tmpFile.Name())
		return
	}
	c.logger.Debugw("applied Kubernetes objects", "resource", r.GetName(), "result", out)
}

// ResourceDeleted is called when a custom resource is created and will generate
// the template files and delete them from Kubernetes
func (c Controller) ResourceDeleted(r *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource deleted", "resource", r.GetName())
	cr := &tmpl.CustomResource{
		Resource: r,
	}
	tmpFile, err := ioutil.TempFile("", "lostromos")
	if err != nil {
		c.logger.Errorw("failed to get tmp file", "error", err)
		return
	}
	err = tmpl.Parse(cr, c.templatePath, tmpFile)
	if err != nil {
		c.logger.Errorw("failed to generate template error", "error", err)
		return
	}
	out, err := c.Client.Delete(tmpFile.Name())
	if err != nil {
		c.logger.Errorw("failed to delete template", "error", err, "result", out)
		c.logger.Debugw("template we want to delete", "template", readFile(tmpFile.Name()), "fileName", tmpFile.Name())
		return
	}
	c.logger.Debugw("deleted Kubernetes objects", "resource", r.GetName(), "result", out)
	metrics.DeletedReleases.Inc()
	metrics.ManagedReleases.Dec()
}

func readFile(filepath string) string {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	return string(content[:])
}
