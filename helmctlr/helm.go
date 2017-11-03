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

package helmctlr

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/wpengine/lostromos/metrics"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/helm/pkg/helm"
)

// Controller is a crwatcher.ResourceController that works with Helm to deploy
// helm charts into K8s providing a CustomResource as value data to the charts
type Controller struct {
	ChartDir    string         // path to dir where the Helm chart is located
	Helm        helm.Interface // Helm for talking with helm
	Namespace   string         // Default namespace to deploy into. If empty it will try to use the namespace from the CustomResource
	ReleaseName string         // Prefix for the helm release name. Will look like ReleaseName-CR_Name
	logger      *zap.SugaredLogger
}

// NewController will return a configured Helm Controller
func NewController(chartDir, ns, rn, host string, logger *zap.SugaredLogger) *Controller {
	if logger == nil {
		// If you don't give us a logger, set logger to a nop logger
		logger = zap.NewNop().Sugar()
	}
	c := &Controller{
		Helm:        helm.NewClient(helm.Host(host)),
		ChartDir:    chartDir,
		Namespace:   ns,
		ReleaseName: rn,
		logger:      logger,
	}
	return c
}

// ResourceAdded is called when a custom resource is created and will kick off a
// help install for the given charts and CR
func (c Controller) ResourceAdded(r *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource added", "resource", r.GetName())
	if err := c.installOrUpdate(r); err != nil {
		metrics.CreateFailures.Inc()
		c.logger.Errorw("failed to create resource", "error", err, "resource", r.GetName())
		return
	}
	metrics.CreatedReleases.Inc()
	metrics.ManagedReleases.Inc()

}

// ResourceDeleted is called when a custom resource is created and will use
// Helm to delete the release. The release is also purged in case in the future
// another CR with the same name is created.
func (c Controller) ResourceDeleted(r *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource deleted", "resource", r.GetName())
	err := c.delete(r)
	if err != nil {
		metrics.DeleteFailures.Inc()
		c.logger.Errorw("failed to delete resource", "error", err, "resource", r.GetName())
		return
	}
	metrics.DeletedReleases.Inc()
	metrics.ManagedReleases.Dec()
}

// ResourceUpdated is called when a custom resource is updated or during a
// resync and will kick off a helm update for the corresponding release
func (c Controller) ResourceUpdated(oldR, newR *unstructured.Unstructured) {
	metrics.TotalEvents.Inc()
	c.logger.Infow("resource updated", "resource", newR.GetName())
	if err := c.installOrUpdate(newR); err != nil {
		metrics.UpdateFailures.Inc()
		c.logger.Errorw("failed to update resource", "error", err, "resource", newR.GetName())
		return
	}
	metrics.UpdatedReleases.Inc()
}

func (c Controller) delete(r *unstructured.Unstructured) error {
	rlsName := c.releaseName(r)
	_, err := c.Helm.DeleteRelease(rlsName, helm.DeletePurge(true))
	return err
}

func (c Controller) installOrUpdate(r *unstructured.Unstructured) error {
	cr, err := c.marshallCR(r)
	if err != nil {
		return err
	}
	rlsName := c.releaseName(r)
	if c.releaseExists(rlsName) {
		_, err = c.Helm.UpdateRelease(
			rlsName,
			c.ChartDir,
			helm.UpdateValueOverrides(cr))
		return err
	}
	_, err = c.Helm.InstallRelease(
		c.ChartDir,
		c.Namespace,
		helm.ReleaseName(rlsName),
		helm.ValueOverrides(cr))
	return err
}

func (c Controller) marshallCR(r *unstructured.Unstructured) ([]byte, error) {
	re := map[string]interface{}{
		"resource": map[string]interface{}{
			"name":      r.GetName(),
			"namespace": r.GetNamespace(),
			"spec":      r.Object["spec"]}}

	return yaml.Marshal(re)
}

func (c Controller) releaseExists(rlsName string) bool {
	r, err := c.Helm.ListReleases(helm.ReleaseListNamespace(c.Namespace), helm.ReleaseListFilter(rlsName))
	if err != nil {
		return false
	}
	for _, i := range r.Releases {
		if i.GetName() == rlsName {
			return true
		}
	}
	return false
}

func (c Controller) releaseName(r *unstructured.Unstructured) string {
	return fmt.Sprintf("%s-%s", c.ReleaseName, r.GetName())
}
