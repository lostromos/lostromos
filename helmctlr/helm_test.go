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
	"io/ioutil"
	"os"
	"testing"

	"github.com/wpengine/lostromos/crwatcher"
	"github.com/wpengine/lostromos/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/tiller/environment"
	internalFake "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
)

const (
	testNamespaceName = "lostromos-test"
	testName          = "dory"
	testReleaseName   = "lostromostest-dory"
	crdGroup          = "stable.nicolerenee.io"
	crdVersion        = "v1alpha1"
	crdKind           = "Character"
)

func newTestResource(t *testing.T, status *crwatcher.CustomResourceStatus) *unstructured.Unstructured {
	resource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       crdKind,
			"apiVersion": crdGroup + "/" + crdVersion,
			"metadata": map[string]interface{}{
				"name":      testName,
				"namespace": testNamespaceName,
			},
			"spec": map[string]interface{}{
				"Name": "Dory",
				"From": "Finding Nemo",
				"By":   "Disney",
			},
		},
	}
	if status != nil {
		statusMap, err := status.ToMap()
		require.NoError(t, err)
		resource.Object["status"] = statusMap
	}
	return resource
}

type FakeController struct {
	Controller
	FakeResourceClient *dynamicFake.FakeResourceClient
}

func newTestController(t *testing.T) *FakeController {
	fakeResourceClient := &dynamicFake.FakeResourceClient{
		Fake:      &k8sTesting.Fake{},
		Resource:  schema.GroupVersionResource{Group: crdGroup, Version: crdVersion, Resource: "characters"},
		Kind:      schema.GroupVersionKind{Group: crdGroup, Version: crdVersion, Kind: crdKind},
		Namespace: testNamespaceName,
	}
	fakeResourceClient.AddReactor("*", "*", func(action k8sTesting.Action) (bool, runtime.Object, error) {
		return true, action.(k8sTesting.UpdateAction).GetObject(), nil
	})

	file, _ := ioutil.TempFile(os.TempDir(), "lostromostest")
	defer os.Remove(file.Name())

	fakeK8sClient := k8sFake.NewSimpleClientset()
	fakeInternalClient := internalFake.NewSimpleClientset()
	ctlr := NewController("../test/data/helm/chart", testNamespaceName, "lostromostest", false, 30, nil, fakeResourceClient, fakeK8sClient, fakeInternalClient)
	ctlr.tillerKubeClient = &environment.PrintingKubeClient{Out: file}
	chartutil.DefaultVersionSet = chartutil.NewVersionSet("v1", "extensions/v1beta1")
	return &FakeController{
		Controller:         *ctlr,
		FakeResourceClient: fakeResourceClient,
	}
}

func assertEqualPhaseAndReason(t *testing.T, a, b *unstructured.Unstructured) {
	aStatus := crwatcher.StatusFor(a)
	bStatus := crwatcher.StatusFor(b)
	assert.EqualValues(t, aStatus.Phase, bStatus.Phase)
	assert.EqualValues(t, aStatus.Reason, bStatus.Reason)
}

func getPromCounterValue(metric string) float64 {
	mf, _ := prometheus.DefaultGatherer.Gather()
	for _, s := range mf {
		if s.GetName() == metric {
			return s.GetMetric()[0].GetCounter().GetValue()
		}
	}
	return 0
}

func getPromGaugeValue(metric string) float64 {
	mf, _ := prometheus.DefaultGatherer.Gather()
	for _, s := range mf {
		if s.GetName() == metric {
			return s.GetMetric()[0].GetGauge().GetValue()
		}
	}
	return 0
}

// Used in assertCounters to mark the expected change in counters
// values default to 0 so you only have to specify the changes
type counterTest struct {
	create    int
	createErr int
	delete    int
	deleteErr int
	update    int
	updateErr int
	events    int
	releases  int
}

func assertCounters(t *testing.T, c counterTest, f func()) {
	metrics.ManagedReleases.Set(float64(10))
	csb := getPromCounterValue("releases_create_total")
	ceb := getPromCounterValue("releases_create_error_total")
	dsb := getPromCounterValue("releases_delete_total")
	deb := getPromCounterValue("releases_delete_error_total")
	usb := getPromCounterValue("releases_update_total")
	ueb := getPromCounterValue("releases_update_error_total")
	eb := getPromCounterValue("releases_events_total")
	rb := getPromGaugeValue("releases_total")

	f()

	csa := getPromCounterValue("releases_create_total")
	cea := getPromCounterValue("releases_create_error_total")
	dsa := getPromCounterValue("releases_delete_total")
	dea := getPromCounterValue("releases_delete_error_total")
	usa := getPromCounterValue("releases_update_total")
	uea := getPromCounterValue("releases_update_error_total")
	ea := getPromCounterValue("releases_events_total")
	ra := getPromGaugeValue("releases_total")
	assert.Equal(t, float64(c.create), csa-csb, "change in releases_create_total incorrect")
	assert.Equal(t, float64(c.createErr), cea-ceb, "change in releases_create_error_total incorrect")
	assert.Equal(t, float64(c.delete), dsa-dsb, "change in releases_delete_total incorrect")
	assert.Equal(t, float64(c.deleteErr), dea-deb, "change in releases_delete_error_total incorrect")
	assert.Equal(t, float64(c.update), usa-usb, "change in releases_update_total incorrect")
	assert.Equal(t, float64(c.updateErr), uea-ueb, "change in releases_update_error_total incorrect")
	assert.Equal(t, float64(c.events), ea-eb, "change in releases_events_total incorrect")
	assert.Equal(t, float64(c.releases), ra-rb, "change in releases_total incorrect")
}

func TestNewControllerSetsNS(t *testing.T) {
	c := NewController("chartDir", "", "release", false, 120, nil, nil, nil, nil)
	assert.Equal(t, "default", c.Namespace, "Namespace should be set to 'default' when not provided")
	assert.Equal(t, "chartDir", c.ChartDir)
	assert.Equal(t, "release", c.ReleaseName)

	c = NewController("chartDir", "my_ns", "release", false, 120, nil, nil, nil, nil)
	assert.Equal(t, "my_ns", c.Namespace, "Namespace should be set to the value provided")
}

func TestResourceAddedHappyPath(t *testing.T) {
	testController := newTestController(t)
	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	assertCounters(t, ct, func() {
		in := newTestResource(t, nil)
		out := newTestResource(t, &crwatcher.CustomResourceStatus{Phase: crwatcher.PhaseApplied, Reason: crwatcher.ReasonApplySuccessful})

		testController.ResourceAdded(in)

		// Update was called
		updateAction, ok := testController.FakeResourceClient.Actions()[0].(k8sTesting.UpdateAction)
		require.True(t, ok)
		assert.True(t, updateAction.Matches("update", testController.FakeResourceClient.Resource.Resource))

		// Update was called with correct new status
		updatedRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(updateAction.GetObject())
		require.NoError(t, err)
		updated := &unstructured.Unstructured{Object: updatedRaw}
		assertEqualPhaseAndReason(t, updated, out)
		require.EqualValues(t, out.Object["spec"], updatedRaw["spec"])
		require.NotNil(t, crwatcher.StatusFor(updated).Release)
	})
}

// Happy path when resource exists...happens on startup
func TestResourceAddedHappyPathExists(t *testing.T) {
	testController := newTestController(t)
	ct := counterTest{
		events:   2,
		create:   2,
		releases: 2,
	}
	assertCounters(t, ct, func() {
		in := newTestResource(t, nil)
		out := newTestResource(t, &crwatcher.CustomResourceStatus{Phase: crwatcher.PhaseApplied, Reason: crwatcher.ReasonApplySuccessful})

		testController.ResourceAdded(in)
		testController.ResourceAdded(in)

		// Update was called
		updateAction, ok := testController.FakeResourceClient.Actions()[1].(k8sTesting.UpdateAction)
		require.True(t, ok)
		assert.True(t, updateAction.Matches("update", testController.FakeResourceClient.Resource.Resource))

		// Update was called with correct new status
		updatedRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(updateAction.GetObject())
		require.NoError(t, err)
		updated := &unstructured.Unstructured{Object: updatedRaw}
		assertEqualPhaseAndReason(t, updated, out)
		require.EqualValues(t, out.Object["spec"], updatedRaw["spec"])
		require.NotNil(t, crwatcher.StatusFor(updated).Release)
	})

	release, err := testController.storage.Last(testReleaseName)
	assert.NoError(t, err)
	assert.Contains(t, release.Manifest, "ownerReferences:\n  - apiVersion: stable.nicolerenee.io")
}

// helm Install returns an error
func TestResourceAddedInstallErrors(t *testing.T) {
	testController := newTestController(t)
	ct := counterTest{
		events:    1,
		createErr: 1,
	}
	chartutil.DefaultVersionSet = chartutil.NewVersionSet("v1")
	assertCounters(t, ct, func() {
		in := newTestResource(t, nil)
		out := newTestResource(t, &crwatcher.CustomResourceStatus{Phase: crwatcher.PhaseFailed, Reason: crwatcher.ReasonApplyFailed})

		testController.ResourceAdded(in)

		// Update was called
		updateAction, ok := testController.FakeResourceClient.Actions()[0].(k8sTesting.UpdateAction)
		require.True(t, ok)
		assert.True(t, updateAction.Matches("update", testController.FakeResourceClient.Resource.Resource))

		// Update was called with correct new status
		updatedRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(updateAction.GetObject())
		require.NoError(t, err)
		updated := &unstructured.Unstructured{Object: updatedRaw}
		assertEqualPhaseAndReason(t, updated, out)
		require.EqualValues(t, out.Object["spec"], updatedRaw["spec"])
		require.Nil(t, crwatcher.StatusFor(updated).Release)
	})
	chartutil.DefaultVersionSet = chartutil.NewVersionSet("v1", "extensions/v1beta1")
}

func TestResourceDeleted(t *testing.T) {
	testController := newTestController(t)
	ct := counterTest{
		events:   1,
		delete:   1,
		releases: -1,
	}
	testController.ResourceAdded(newTestResource(t, nil))
	assertCounters(t, ct, func() {
		testController.ResourceDeleted(newTestResource(t, nil))
	})
}

func TestResourceDeletedWhenDeleteFails(t *testing.T) {
	testController := newTestController(t)
	ct := counterTest{
		events:    1,
		deleteErr: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceDeleted(newTestResource(t, nil))
	})
}

func TestResourceUpdatedHappyPath(t *testing.T) {
	testController := newTestController(t)
	ct := counterTest{
		events: 1,
		update: 1,
	}
	assertCounters(t, ct, func() {
		in := newTestResource(t, nil)
		updatedIn := newTestResource(t, nil)
		out := newTestResource(t, &crwatcher.CustomResourceStatus{Phase: crwatcher.PhaseApplied, Reason: crwatcher.ReasonApplySuccessful})

		testController.ResourceUpdated(in, updatedIn)

		// Update was called
		updateAction, ok := testController.FakeResourceClient.Actions()[0].(k8sTesting.UpdateAction)
		require.True(t, ok)
		assert.True(t, updateAction.Matches("update", testController.FakeResourceClient.Resource.Resource))

		// Update was called with correct new status
		updatedRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(updateAction.GetObject())
		require.NoError(t, err)
		updated := &unstructured.Unstructured{Object: updatedRaw}
		assertEqualPhaseAndReason(t, updated, out)
		require.EqualValues(t, out.Object["spec"], updatedRaw["spec"])
		require.NotNil(t, crwatcher.StatusFor(updated).Release)
	})
}

// Happy path when resource exists...happens on startup
func TestResourceUpdatedHappyPathExists(t *testing.T) {
	testController := newTestController(t)
	ct := counterTest{
		events: 2,
		update: 2,
	}
	assertCounters(t, ct, func() {
		in := newTestResource(t, nil)
		updatedIn := newTestResource(t, nil)
		out := newTestResource(t, &crwatcher.CustomResourceStatus{Phase: crwatcher.PhaseApplied, Reason: crwatcher.ReasonApplySuccessful})

		testController.ResourceUpdated(in, updatedIn)
		testController.ResourceUpdated(in, updatedIn)

		// Update was called
		updateAction, ok := testController.FakeResourceClient.Actions()[1].(k8sTesting.UpdateAction)
		require.True(t, ok)
		assert.True(t, updateAction.Matches("update", testController.FakeResourceClient.Resource.Resource))

		// Update was called with correct new status
		updatedRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(updateAction.GetObject())
		require.NoError(t, err)
		updated := &unstructured.Unstructured{Object: updatedRaw}
		assertEqualPhaseAndReason(t, updated, out)
		require.EqualValues(t, out.Object["spec"], updatedRaw["spec"])
		require.NotNil(t, crwatcher.StatusFor(updated).Release)
	})
}

// helm Install returns an error
func TestResourceUpdatedInstallErrors(t *testing.T) {
	testController := newTestController(t)

	ct := counterTest{
		events:    1,
		updateErr: 1,
	}
	assertCounters(t, ct, func() {
		chartutil.DefaultVersionSet = chartutil.NewVersionSet("v1")
		testController.ResourceUpdated(newTestResource(t, nil), newTestResource(t, nil))
		chartutil.DefaultVersionSet = chartutil.NewVersionSet("v1", "extensions/v1beta1")
	})
}

// if helm controller is used to load a resource, the release state should be read from the CR
// this ensures a consistant view of releases through restarts or multiple deployments
func TestResourceUpdatedReusesStoredRelease(t *testing.T) {
	testController1 := newTestController(t)

	in := newTestResource(t, nil)
	out := newTestResource(t, &crwatcher.CustomResourceStatus{Phase: crwatcher.PhaseApplied, Reason: crwatcher.ReasonApplySuccessful})

	testController1.ResourceAdded(in)

	// Update was called
	updateAction, ok := testController1.FakeResourceClient.Actions()[0].(k8sTesting.UpdateAction)
	require.True(t, ok)
	assert.True(t, updateAction.Matches("update", testController1.FakeResourceClient.Resource.Resource))

	// Update was called with correct new status
	updatedRaw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(updateAction.GetObject())
	require.NoError(t, err)
	updated := &unstructured.Unstructured{Object: updatedRaw}
	assertEqualPhaseAndReason(t, updated, out)
	require.EqualValues(t, out.Object["spec"], updatedRaw["spec"])

	originalStatus := crwatcher.StatusFor(updated)
	require.NotNil(t, originalStatus.Release)

	// new controller to simulate restart or a separate node
	testController2 := newTestController(t)
	testController2.ResourceAdded(updated)

	// Added resource with a release status should've updated the internal storage with the release
	parsedRelease, err := testController2.storage.Last(testReleaseName)
	require.NoError(t, err)
	require.NotNil(t, parsedRelease)
}
