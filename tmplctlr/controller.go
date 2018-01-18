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
	"os"
	"path/filepath"

	"github.com/wpengine/lostromos/metrics"
	"github.com/wpengine/lostromos/tmpl"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// Controller implements a valid crwatcher.ResourceController that will manage
// resources in kubernetes based on the provided template files.
type Controller struct {
	templatePath string     //path to dir where templates are located
	Client       KubeClient //client for talking with kubernetes
	logger       *zap.SugaredLogger
	kubeClient   *dynamic.Client
}

// NewController will return a configured Controller
func NewController(tmplDir string, kubeCfg string, logger *zap.SugaredLogger, kubeClient *dynamic.Client) *Controller {
	if logger == nil {
		// If you don't give us a logger, set logger to a nop logger
		logger = zap.NewNop().Sugar()
	}
	c := &Controller{
		Client:       &Kubectl{ConfigFile: kubeCfg},
		templatePath: filepath.Join(tmplDir, "*.tmpl"),
		logger:       logger,
		kubeClient:   kubeClient,
	}
	return c
}

// ResourceAdded is called when a custom resource is created and will generate
// the template files and apply them to Kubernetes
func (c Controller) ResourceAdded(r *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource added", "resource", r.GetName())
	out, err := c.apply(r)
	if err != nil {
		c.logger.Errorw("failed to add resource", "resource", r.GetName(), "error", err, "cmdOutput", out)
		metrics.CreateFailures.Inc()
		return
	}
	metrics.CreatedReleases.Inc()
	metrics.ManagedReleases.Inc()
}

// ResourceUpdated is called when a custom resource is updated or during a
// resync and will generate the template files and apply them to Kubernetes
func (c Controller) ResourceUpdated(oldR, newR *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource updated", "resource", newR.GetName())
	out, err := c.apply(newR)
	if err != nil {
		c.logger.Errorw("failed to update resource", "resource", newR.GetName(), "error", err, "cmdOutput", out)
		metrics.UpdateFailures.Inc()
		return
	}
	metrics.UpdatedReleases.Inc()
}

// ResourceDeleted is called when a custom resource is created and will generate
// the template files and delete them from Kubernetes
func (c Controller) ResourceDeleted(r *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource deleted", "resource", r.GetName())
	out, err := c.delete(r)
	if err != nil {
		c.logger.Errorw("failed to delete resource", "resource", r.GetName(), "error", err, "cmdOutput", out)
		metrics.DeleteFailures.Inc()
		return
	}
	metrics.DeletedReleases.Inc()
	metrics.ManagedReleases.Dec()
}

func (c Controller) apply(r *unstructured.Unstructured) (output string, err error) {
	tmpFile, err := c.buildTemplate(r)
	if err != nil {
		return "", err
	}
	return c.Client.Apply(tmpFile.Name())
}

func (c Controller) delete(r *unstructured.Unstructured) (output string, err error) {
	tmpFile, err := c.buildTemplate(r)
	if err != nil {
		return "", err
	}
	return c.Client.Delete(tmpFile.Name())
}

func (c Controller) buildTemplate(r *unstructured.Unstructured) (tmpFile *os.File, err error) {
	cr := &tmpl.CustomResource{
		Resource: r,
	}
	tmpFile, err = ioutil.TempFile("", "lostromos")
	if err != nil {
		return tmpFile, err
	}
	err = tmpl.Parse(cr, c.templatePath, tmpFile)
	return tmpFile, err
}
