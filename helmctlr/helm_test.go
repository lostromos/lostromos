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
	"time"

	"os"
	"path/filepath"

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
	testController  = helmctlr.NewController("../test/data/chart", "lostromos-test", "lostromostest", "0", false, 30, nil)
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
	testRemoteRepoResource = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "dory",
				"annotations": map[string]interface{}{
					"chart": "test/helloworld:0.1.0",
				},
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
	create        int
	createErr     int
	delete        int
	deleteErr     int
	update        int
	updateErr     int
	events        int
	eventQueue    int
	releases      int
	remoteRepo    int
	remoteRepoErr int
}

func timestampTestMap() map[string]func(float64, float64) bool {
	timestamps := []string{"releases_last_create_timestamp_utc_seconds",
		"releases_last_update_timestamp_utc_seconds",
		"releases_last_delete_timestamp_utc_seconds"}
	timestampMetrics := make(map[string]func(float64, float64) bool)
	for _, t := range timestamps {
		timestampMetrics[t] = lessThan
	}
	return timestampMetrics
}

func assertMetrics(t *testing.T, c counterTest, f func(), tsMap map[string]func(float64, float64) bool) {
	metrics.ManagedReleases.Set(float64(10))
	csb := getPromCounterValue("releases_create_total")
	ceb := getPromCounterValue("releases_create_error_total")
	dsb := getPromCounterValue("releases_delete_total")
	deb := getPromCounterValue("releases_delete_error_total")
	usb := getPromCounterValue("releases_update_total")
	ueb := getPromCounterValue("releases_update_error_total")
	eb := getPromCounterValue("releases_events_total")
	rb := getPromGaugeValue("releases_total")
	rrb := getPromCounterValue("releases_remote_repo_total")
	rreb := getPromCounterValue("releases_remote_repo_error_total")

	currTimestamp := float64(time.Now().UTC().UnixNano()) / 1000000000
	f()

	// assert counters
	csa := getPromCounterValue("releases_create_total")
	cea := getPromCounterValue("releases_create_error_total")
	dsa := getPromCounterValue("releases_delete_total")
	dea := getPromCounterValue("releases_delete_error_total")
	usa := getPromCounterValue("releases_update_total")
	uea := getPromCounterValue("releases_update_error_total")
	ea := getPromCounterValue("releases_events_total")
	ra := getPromGaugeValue("releases_total")
	rra := getPromCounterValue("releases_remote_repo_total")
	rrea := getPromCounterValue("releases_remote_repo_error_total")
	assert.Equal(t, float64(c.create), csa-csb, "change in releases_create_total incorrect")
	assert.Equal(t, float64(c.createErr), cea-ceb, "change in releases_create_error_total incorrect")
	assert.Equal(t, float64(c.delete), dsa-dsb, "change in releases_delete_total incorrect")
	assert.Equal(t, float64(c.deleteErr), dea-deb, "change in releases_delete_error_total incorrect")
	assert.Equal(t, float64(c.update), usa-usb, "change in releases_update_total incorrect")
	assert.Equal(t, float64(c.updateErr), uea-ueb, "change in releases_update_error_total incorrect")
	assert.Equal(t, float64(c.events), ea-eb, "change in releases_events_total incorrect")
	assert.Equal(t, float64(c.releases), ra-rb, "change in releases_total incorrect")
	assert.Equal(t, float64(c.eventQueue), 0.0, "change in releases_event_queue incorrect")
	assert.Equal(t, float64(c.remoteRepo), rra-rrb, "change in remote_repo_total incorrect")
	assert.Equal(t, float64(c.remoteRepoErr), rrea-rreb, "change in remote_repo_error_total incorrect")

	//assert timestamps
	for metric, f := range tsMap {
		assertTimestamp(t, metric, f, currTimestamp)
	}
}

func assertTimestamp(t *testing.T, metric_name string, f func(float64, float64) bool, ts float64) {
	metric_ts := getPromGaugeValue(metric_name)
	assert.True(t, f(metric_ts, ts), "%s timestamp did not update correctly", metric_name)
}

func greaterThan(a float64, b float64) bool {
	return a > b
}

func lessThan(a float64, b float64) bool {
	return a < b
}

func TestNewControllerSetsNS(t *testing.T) {
	c := helmctlr.NewController("chartDir", "", "release", "127.0.0.3:4321", false, 120, nil)
	assert.Equal(t, "default", c.Namespace, "Namespace should be set to 'default' when not provided")
	assert.Equal(t, "chartDir", c.ChartPath)
	assert.Equal(t, "release", c.ReleaseName)

	c = helmctlr.NewController("chartDir", "my_ns", "release", "127.0.0.3:4321", false, 120, nil)
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
	mockHelm.EXPECT().InstallRelease(testController.ChartPath, testController.Namespace, installOpts...)

	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_create_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceAdded(testResource) }, tsExpected)
}

func TestResourceAddedNoPriorReleaseHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(nil, nil)
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartPath, testController.Namespace, installOpts...)

	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_create_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceAdded(testResource) }, tsExpected)
}

func TestResourceRemoteRepoAddedHappyPath(t *testing.T) {
	repoSrv := SetupMockServer(t)
	defer repoSrv.Cleanup()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(&services.ListReleasesResponse{}, nil)
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	chartFilename := "chart-0.1.0.tgz"
	chartPath := filepath.Join(os.TempDir(), "test/helloworld", helmctlr.Hash(chartFilename), chartFilename)
	mockHelm.EXPECT().InstallRelease(chartPath, testController.Namespace, installOpts...)

	ct := counterTest{
		events:     1,
		create:     1,
		releases:   1,
		remoteRepo: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_create_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceAdded(testRemoteRepoResource) }, tsExpected)
}

func TestResourceRemoteRepoAddedFailureCase(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm

	ct := counterTest{
		events:        1,
		createErr:     1,
		remoteRepoErr: 1,
	}

	assertMetrics(t, ct, func() { testController.ResourceAdded(testRemoteRepoResource) }, timestampTestMap())
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
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartPath, opts...)

	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_create_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceAdded(testResource) }, tsExpected)
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
	mockHelm.EXPECT().InstallRelease(testController.ChartPath, testController.Namespace, installOpts...)

	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_create_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceAdded(testResource) }, tsExpected)
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
	mockHelm.EXPECT().InstallRelease(testController.ChartPath, testController.Namespace, installOpts...).Return(nil, errors.New("install failed"))

	ct := counterTest{
		events:    1,
		createErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { testController.ResourceAdded(testResource) }, tsExpected)
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
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartPath, opts...).Return(nil, errors.New("install failed"))

	ct := counterTest{
		events:    1,
		createErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { testController.ResourceAdded(testResource) }, tsExpected)
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
	tsExpected := timestampTestMap()
	tsExpected["releases_last_delete_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceDeleted(testResource) }, tsExpected)
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
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { testController.ResourceDeleted(testResource) }, tsExpected)
}

func TestResourceUpdatedHappyPath(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockHelm := NewMockInterface(mockCtrl)
	testController.Helm = mockHelm
	listOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().ListReleases(listOpts...).Return(&services.ListReleasesResponse{}, nil)
	installOpts := []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}
	mockHelm.EXPECT().InstallRelease(testController.ChartPath, testController.Namespace, installOpts...)

	ct := counterTest{
		events: 1,
		update: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_update_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceUpdated(testResource, testResource) }, tsExpected)
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
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartPath, opts...)

	ct := counterTest{
		events: 1,
		update: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_update_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceUpdated(testResource, testResource) }, tsExpected)
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
	mockHelm.EXPECT().InstallRelease(testController.ChartPath, testController.Namespace, installOpts...)

	ct := counterTest{
		events: 1,
		update: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_update_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { testController.ResourceUpdated(testResource, testResource) }, tsExpected)
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
	mockHelm.EXPECT().InstallRelease(testController.ChartPath, testController.Namespace, installOpts...).Return(nil, errors.New("install failed"))

	ct := counterTest{
		events:    1,
		updateErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { testController.ResourceUpdated(testResource, testResource) }, tsExpected)
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
	mockHelm.EXPECT().UpdateRelease(testReleaseName, testController.ChartPath, opts...).Return(nil, errors.New("install failed"))

	ct := counterTest{
		events:    1,
		updateErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { testController.ResourceUpdated(testResource, testResource) }, tsExpected)
}
