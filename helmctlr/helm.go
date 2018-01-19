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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
)

var defaultNS = "default"

// Controller is a crwatcher.ResourceController that works with Helm to deploy
// helm charts into K8s providing a CustomResource as value data to the charts
type Controller struct {
	ChartDir       string         // path to dir where the Helm chart is located
	Helm           helm.Interface // Helm for talking with helm
	Namespace      string         // Default namespace to deploy into. If empty it will default to "default"
	ReleaseName    string         // Prefix for the helm release name. Will look like ReleaseName-CR_Name
	Wait           bool           // Whether or not to wait for resources during Update and Install before marking a release successful
	WaitTimeout    int64          // time in seconds to wait for kubernetes resources to be created before marking a release successful
	logger         *zap.SugaredLogger
	kubeClient     kubernetes.Interface
	resourceClient dynamic.ResourceInterface
}

// NewController will return a configured Helm Controller
func NewController(chartDir, ns, rn, host string, wait bool, waitto int64, logger *zap.SugaredLogger, resourceClient dynamic.ResourceInterface, kubeClient kubernetes.Interface) *Controller {
	if logger == nil {
		// If you don't give us a logger, set logger to a nop logger
		logger = zap.NewNop().Sugar()
	}
	if ns == "" {
		ns = defaultNS
	}
	c := &Controller{
		Helm:           helm.NewClient(helm.Host(host)),
		ChartDir:       chartDir,
		Namespace:      ns,
		ReleaseName:    rn,
		Wait:           wait,
		WaitTimeout:    waitto,
		resourceClient: resourceClient,
		kubeClient:     kubeClient,
		logger:         logger,
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
			helm.UpdateValueOverrides(cr),
			helm.UpgradeWait(c.Wait),
			helm.UpgradeTimeout(c.WaitTimeout))
		return err
	}
	_, err = c.Helm.InstallRelease(
		c.ChartDir,
		c.Namespace,
		helm.ReleaseName(rlsName),
		helm.ValueOverrides(cr),
		helm.InstallWait(c.Wait),
		helm.InstallTimeout(c.WaitTimeout))
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
	statuses := []release.Status_Code{
		release.Status_UNKNOWN,
		release.Status_DEPLOYED,
		release.Status_DELETED,
		release.Status_DELETING,
		release.Status_FAILED,
		release.Status_PENDING_INSTALL,
		release.Status_PENDING_UPGRADE,
		release.Status_PENDING_ROLLBACK,
	}
	r, err := c.Helm.ListReleases(
		helm.ReleaseListNamespace(c.Namespace),
		helm.ReleaseListFilter(rlsName),
		helm.ReleaseListStatuses(statuses),
	)
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
