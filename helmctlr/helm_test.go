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

package helmctlr_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/wpengine/lostromos/helmctlr"
	"github.com/wpengine/lostromos/metrics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
)

var (
	testController  = helmctlr.NewController("../test/data/chart", "lostromos-test", "lostromostest", "0", false, 30, nil, nil, nil)
	testReleaseName = "lostromostest-dory"
	testResource    = &unstructured.Unstructured{
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
)

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
	c := helmctlr.NewController("chartDir", "", "release", "127.0.0.3:4321", false, 120, nil, nil, nil)
	assert.Equal(t, "default", c.Namespace, "Namespace should be set to 'default' when not provided")
	assert.Equal(t, "chartDir", c.ChartDir)
	assert.Equal(t, "release", c.ReleaseName)

	c = helmctlr.NewController("chartDir", "my_ns", "release", "127.0.0.3:4321", false, 120, nil, nil, nil)
	assert.Equal(t, "my_ns", c.Namespace, "Namespace should be set to the value provided")
}

func TestResourceAddedHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(&services.ListReleasesResponse{}, nil)
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartDir, testController.Namespace, installOpts...)

	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceAdded(testResource)
	})
}

// Happy path when resource exists...happens on startup
func TestResourceAddedHappyPathExists(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	res := &services.ListReleasesResponse{
		Count: int64(1),
		Releases: []*release.Release{
			{
				Name: testReleaseName,
			}}}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(res, nil)
	opts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartDir, opts...)
	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceAdded(testResource)
	})
}

// List returns an error but install still works
func TestResourceAddedListErrorStillSuccessful(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(nil, errors.New("Broken"))
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartDir, testController.Namespace, installOpts...)

	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceAdded(testResource)
	})
}

// helm Install returns an error
func TestResourceAddedInstallErrors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(&services.ListReleasesResponse{}, nil)
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartDir, testController.Namespace, installOpts...).Return(nil, errors.New("install failed"))

	ct := counterTest{
		events:    1,
		createErr: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceAdded(testResource)
	})
}

// helm update returns an error
func TestResourceAddedUpdateErrors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	res := &services.ListReleasesResponse{
		Count: int64(1),
		Releases: []*release.Release{
			{
				Name: testReleaseName,
			}}}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(res, nil)
	opts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartDir, opts...).Return(nil, errors.New("install failed"))
	ct := counterTest{
		events:    1,
		createErr: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceAdded(testResource)
	})
}

func TestResourceDeleted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	deleteOpts := []interface{}{gomock.Any()}
	mockHelm.EXPECT().DeleteRelease(testReleaseName, deleteOpts...)

	ct := counterTest{
		events:   1,
		delete:   1,
		releases: -1,
	}

	assertCounters(t, ct, func() {
		testController.ResourceDeleted(testResource)
	})
}

func TestResourceDeletedWhenDeleteFails(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	deleteOpts := []interface{}{gomock.Any()}
	mockHelm.EXPECT().DeleteRelease(testReleaseName, deleteOpts...).Return(nil, errors.New("delete failed"))
	ct := counterTest{
		events:    1,
		deleteErr: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceDeleted(testResource)
	})
}

func TestResourceUpdatedHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(&services.ListReleasesResponse{}, nil)
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartDir, testController.Namespace, installOpts...)

	ct := counterTest{
		events: 1,
		update: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceUpdated(testResource, testResource)
	})
}

// Happy path when resource exists...happens on startup
func TestResourceUpdatedHappyPathExists(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	res := &services.ListReleasesResponse{
		Count: int64(1),
		Releases: []*release.Release{
			{
				Name: testReleaseName,
			}}}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(res, nil)
	opts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartDir, opts...)
	ct := counterTest{
		events: 1,
		update: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceUpdated(testResource, testResource)
	})
}

// List returns an error but install still works
func TestResourceUpdatedListErrorStillSuccessful(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(nil, errors.New("Broken"))
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartDir, testController.Namespace, installOpts...)

	ct := counterTest{
		events: 1,
		update: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceUpdated(testResource, testResource)
	})
}

// helm Install returns an error
func TestResourceUpdatedInstallErrors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(&services.ListReleasesResponse{}, nil)
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartDir, testController.Namespace, installOpts...).Return(nil, errors.New("install failed"))

	ct := counterTest{
		events:    1,
		updateErr: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceUpdated(testResource, testResource)
	})
}

// helm update returns an error
func TestResourceUpdatedUpdateErrors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	res := &services.ListReleasesResponse{
		Count: int64(1),
		Releases: []*release.Release{
			{
				Name: testReleaseName,
			}}}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(res, nil)
	opts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartDir, opts...).Return(nil, errors.New("install failed"))
	ct := counterTest{
		events:    1,
		updateErr: 1,
	}
	assertCounters(t, ct, func() {
		testController.ResourceUpdated(testResource, testResource)
	})
}
