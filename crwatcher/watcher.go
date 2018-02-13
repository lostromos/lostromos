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

package crwatcher

import (
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

// Config provides config for a CRD Watcher
type Config struct {
	Group      string        // API Group of the CRD
	Namespace  string        // namespace of the CRD
	Version    string        // version of the CRD
	PluralName string        // plural name of the CRD
	Filter     string        // Optional disregard resources that don't have an annotation key matching this filter
	Resync     time.Duration // How often existing CRs should be resynced (marked as updated)
}

// CRWatcher thing that watches
type CRWatcher struct {
	Config     *Config
	resource   dynamic.ResourceInterface
	handler    cache.ResourceEventHandlerFuncs
	store      cache.Store
	controller cache.Controller
	logger     ErrorLogger
}

// ResourceController exposes the functionality of a controller that
// will handle callbacks for events that happen to the Custom Resource being
// monitored. The events are informational only, so you can't return an
// error.
//  * ResourceAdded is called when an object is added.
//  * ResourceUpdated is called when an object is modified. Note that
//      oldResource is the last known state of the object-- it is possible that
//      several changes were combined together, so you can't use this to see
//      every single change. ResourceUpdated is also called when a re-list
//      happens, and it will get called even if nothing changed. This is useful
//      for periodically evaluating or syncing something.
//  * ResourceDeleted will get the final state of the item if it is known,
//      otherwise it will get an object of type DeletedFinalStateUnknown. This
//      can happen if the watch is closed and misses the delete event and we
//      don't notice the deletion until the subsequent re-list.
type ResourceController interface {
	ResourceAdded(resource *unstructured.Unstructured)
	ResourceUpdated(oldResource, newResource *unstructured.Unstructured)
	ResourceDeleted(resource *unstructured.Unstructured)
}

// ErrorLogger will receive any error messages from the kubernetes client
type ErrorLogger interface {
	Error(err error)
}

// NewCRWatcher builds a CRWatcher
func NewCRWatcher(cfg *Config, dc dynamic.Interface, rc ResourceController, l ErrorLogger) (*CRWatcher, error) {
	cw := &CRWatcher{
		Config: cfg,
		logger: l,
	}

	cw.setupResource(dc)
	cw.setupHandler(rc)
	cw.setupController()
	cw.setupRuntimeLogging()
	return cw, nil
}

func (cw *CRWatcher) setupRuntimeLogging() {
	if cw.logger != nil {
		utilruntime.ErrorHandlers = []func(error){
			cw.logKubeError,
		}
	}
}

func (cw *CRWatcher) logKubeError(err error) {
	cw.logger.Error(err)
}

func (cw *CRWatcher) setupHandler(con ResourceController) {
	cw.handler = cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			r := obj.(*unstructured.Unstructured)
			if cw.passesFiltering(r) {
				con.ResourceAdded(r)
			}
		},
		DeleteFunc: func(obj interface{}) {
			r := obj.(*unstructured.Unstructured)
			if cw.passesFiltering(r) {
				con.ResourceDeleted(r)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldR := oldObj.(*unstructured.Unstructured)
			newR := newObj.(*unstructured.Unstructured)
			cw.update(con, oldR, newR)
		},
	}
}

// update sends an appropriate notification to the controller based on filtering outcomes of the old and new state of a
// resource.
//
// If no filter is configured or both states of the resource pass filtering, send an update to the controller.
// If the new state passes filtering and the old state does not, send an add notification to the controller.
// If the old state passes filtering and the new state does not, send a delete notification to the controller.
// If neither state passes filtering, ignore.
//
func (cw *CRWatcher) update(con ResourceController, oldR *unstructured.Unstructured, newR *unstructured.Unstructured) {
	if cw.passesFiltering(newR) {
		if cw.passesFiltering(oldR) {
			con.ResourceUpdated(oldR, newR)
			return
		}
		con.ResourceAdded(newR)
	} else if cw.passesFiltering(oldR) {
		con.ResourceDeleted(oldR)
	}
}

func (cw *CRWatcher) setupResource(dc dynamic.Interface) {
	apiResource := &metav1.APIResource{
		Name:       cw.Config.PluralName,
		Namespaced: cw.Config.Namespace != metav1.NamespaceNone,
	}
	cw.resource = dc.Resource(apiResource, cw.Config.Namespace)
}

func (cw *CRWatcher) setupController() {
	listFunc := func(opts metav1.ListOptions) (runtime.Object, error) {
		return cw.resource.List(opts)
	}
	watchFunc := func(opts metav1.ListOptions) (watch.Interface, error) {
		return cw.resource.Watch(opts)
	}
	lw := &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
	cw.store, cw.controller = cache.NewInformer(
		lw,
		&unstructured.Unstructured{},
		cw.Config.Resync,
		cw.handler,
	)
}

// passesFiltering checks to see if we are using an opt in filter (if not, then return true), and if so returns whether we
// have an annotation matching the given filter.
func (cw *CRWatcher) passesFiltering(r *unstructured.Unstructured) bool {
	if cw.Config.Filter == "" {
		return true
	}

	annotations := r.GetAnnotations()
	if annotations == nil {
		return false
	}

	_, ok := annotations[cw.Config.Filter]
	return ok
}

// Watch will be called to begin watching the configured custom resource. All
// events will be passed back to the ResourceController
func (cw *CRWatcher) Watch(stopCh <-chan struct{}) error {
	if cw.controller == nil {
		return errors.New("the CRWatcher has not been initialized")
	}
	cw.controller.Run(stopCh)
	return nil
}
