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
	"context"
	"fmt"

	"github.com/ghodss/yaml"
	crw "github.com/wpengine/lostromos/crwatcher"
	"github.com/wpengine/lostromos/metrics"
	"go.uber.org/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/kube"
	cpb "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/storage"
	"k8s.io/helm/pkg/storage/driver"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/tiller/environment"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

var defaultNS = "default"

// Controller is a crwatcher.ResourceController that works with Helm to deploy
// helm charts into K8s providing a CustomResource as value data to the charts
type Controller struct {
	ChartDir         string // path to dir where the Helm chart is located
	Namespace        string // Default namespace to deploy into. If empty it will default to "default"
	ReleaseName      string // Prefix for the helm release name. Will look like ReleaseName-CR_Name
	Wait             bool   // Whether or not to wait for resources during Update and Install before marking a release successful
	WaitTimeout      int64  // time in seconds to wait for kubernetes resources to be created before marking a release successful
	logger           *zap.SugaredLogger
	kubeClient       kubernetes.Interface
	resourceClient   dynamic.ResourceInterface
	internalClient   internalclientset.Interface // tiller uses internalclientset instead of client-go
	tillerKubeClient environment.KubeClient      // tiller-specific kubernnetes client
	storage          *storage.Storage
}

// NewController will return a configured Helm Controller
func NewController(chartDir, ns, rn string, wait bool, waitto int64, logger *zap.SugaredLogger, resourceClient dynamic.ResourceInterface, kubeClient kubernetes.Interface, internalClient internalclientset.Interface) *Controller {
	if logger == nil {
		// If you don't give us a logger, set logger to a nop logger
		logger = zap.NewNop().Sugar()
	}
	if ns == "" {
		ns = defaultNS
	}
	c := &Controller{
		ChartDir:         chartDir,
		Namespace:        ns,
		ReleaseName:      rn,
		Wait:             wait,
		WaitTimeout:      waitto,
		resourceClient:   resourceClient,
		kubeClient:       kubeClient,
		internalClient:   internalClient,
		logger:           logger,
		storage:          storage.Init(driver.NewMemory()),
		tillerKubeClient: kube.New(nil),
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
		c.updateCRStatus(r, crw.PhaseFailed, crw.ReasonApplyFailed, err.Error())
		return
	}
	metrics.CreatedReleases.Inc()
	metrics.ManagedReleases.Inc()

	c.updateCRStatus(r, crw.PhaseApplied, crw.ReasonApplySuccessful, "")
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
		c.updateCRStatus(newR, crw.PhaseFailed, crw.ReasonApplyFailed, err.Error())
		return
	}
	metrics.UpdatedReleases.Inc()

	c.updateCRStatus(newR, crw.PhaseApplied, crw.ReasonApplySuccessful, "")
}

func (c Controller) delete(r *unstructured.Unstructured) error {
	rlsName := c.releaseName(r)
	tiller := c.tillerRendererForCR(r)

	// TODO: useful status from response?
	_, err := tiller.UninstallRelease(context.TODO(), &services.UninstallReleaseRequest{
		Name:  rlsName,
		Purge: true,
	})
	return err
}

func (c Controller) installOrUpdate(r *unstructured.Unstructured) error {
	cr, err := c.marshallCR(r)
	if err != nil {
		return err
	}

	rlsName := c.releaseName(r)

	tiller := c.tillerRendererForCR(r)

	// load chart
	chart, err := chartutil.LoadDir(c.ChartDir)
	if err != nil {
		return err
	}
	release, err := c.storage.Last(rlsName)
	if err != nil || release == nil {
		//install release
		installReq := &services.InstallReleaseRequest{
			Namespace: r.GetNamespace(),
			Name:      rlsName,
			Chart:     chart,
			Values:    &cpb.Config{Raw: string(cr)},
			Wait:      c.Wait,
			Timeout:   c.WaitTimeout,
		}
		//TODO: useful status from response?
		_, err := tiller.InstallRelease(context.TODO(), installReq)
		if err != nil {
			return err
		}
	} else {
		//update release
		updateReq := &services.UpdateReleaseRequest{
			Name:    rlsName,
			Chart:   chart,
			Values:  &cpb.Config{Raw: string(cr)},
			Wait:    c.Wait,
			Timeout: c.WaitTimeout,
		}
		//TODO: useful status from response?
		_, err := tiller.UpdateRelease(context.TODO(), updateReq)
		if err != nil {
			return err
		}
	}
	return nil
}

// tillerRendererForCR creates a ReleaseServer configured with a rendering engine that adds ownerrefs to rendered assets
// based on the CR
func (c Controller) tillerRendererForCR(r *unstructured.Unstructured) *tiller.ReleaseServer {
	controllerRef := metav1.NewControllerRef(r, r.GroupVersionKind())
	ownerRefs := []metav1.OwnerReference{
		*controllerRef,
	}
	baseEngine := engine.New()
	e := NewOwnerRefEngine(baseEngine, ownerRefs)
	var ey environment.EngineYard = map[string]environment.Engine{
		environment.GoTplEngine: e,
	}
	env := &environment.Environment{
		EngineYard: ey,
		Releases:   c.storage,
		KubeClient: c.tillerKubeClient,
	}
	return tiller.NewReleaseServer(env, c.internalClient, false)
}

func (c Controller) marshallCR(r *unstructured.Unstructured) ([]byte, error) {
	re := map[string]interface{}{
		"resource": map[string]interface{}{
			"name":      r.GetName(),
			"namespace": r.GetNamespace(),
			"spec":      r.Object["spec"]}}

	return yaml.Marshal(re)
}

func (c Controller) releaseName(r *unstructured.Unstructured) string {
	return fmt.Sprintf("%s-%s", c.ReleaseName, r.GetName())
}

func (c Controller) updateCRStatus(r *unstructured.Unstructured, phase crw.ResourcePhase, reason crw.ConditionReason, message string) (*unstructured.Unstructured, error) {
	updatedResource := r.DeepCopy()
	updatedResource.Object["status"] = crw.SetPhase(crw.StatusFor(r), phase, reason, message)
	return c.resourceClient.Update(updatedResource)
}
