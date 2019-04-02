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

package tmplctlr_test

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/lostromos/lostromos/metrics"
	"github.com/lostromos/lostromos/tmplctlr"
)

var (
	testResource = &unstructured.Unstructured{
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

	testTemplates = []testFile{
		// T0.tmpl is a plain template file that just invokes T1.
		{"0_base.tmpl", `--- {{template "file1.tmpl" . }}`},
		// T1.tmpl defines a template, T1 that invokes T2.
		{"file1.tmpl", `name: {{ .GetField "metadata" "name"  }}-configmap`},
	}

	testBadTemplates = []testFile{
		{"base", `--- {{template "not there.tmpl" . }}`},
	}
)

// templateFile defines the contents of a template to be stored in a file, for testing.
type testFile struct {
	name     string
	contents string
}

func createTestDir(files []testFile) string {
	dir, err := ioutil.TempDir("", "template")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		f, err := os.Create(filepath.Join(dir, file.name))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err = io.WriteString(f, file.contents)
		if err != nil {
			log.Fatal(err)
		}
	}
	return dir
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

// Used in assertMetrics to mark the expected change in counters
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

// Used in assertMetrics to mark the expected change in timestamps
// values default function is 'less than' current timestamp
func timestampTestMap() map[string]func(float64, float64) bool {
	timestamps := []string{
		"releases_last_create_timestamp_utc_seconds",
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
	assert.Equal(t, float64(c.create), csa-csb, "change in releases_create_total incorrect")
	assert.Equal(t, float64(c.createErr), cea-ceb, "change in releases_create_error_total incorrect")
	assert.Equal(t, float64(c.delete), dsa-dsb, "change in releases_delete_total incorrect")
	assert.Equal(t, float64(c.deleteErr), dea-deb, "change in releases_delete_error_total incorrect")
	assert.Equal(t, float64(c.update), usa-usb, "change in releases_update_total incorrect")
	assert.Equal(t, float64(c.updateErr), uea-ueb, "change in releases_update_error_total incorrect")
	assert.Equal(t, float64(c.events), ea-eb, "change in releases_events_total incorrect")
	assert.Equal(t, float64(c.releases), ra-rb, "change in releases_total incorrect")

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

func TestResourceAddedHappyPath(t *testing.T) {
	dir := createTestDir(testTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	mockKube.EXPECT().Apply(gomock.Any())

	ct := counterTest{
		events:   1,
		create:   1,
		releases: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_create_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { c.ResourceAdded(testResource) }, tsExpected)
}

func TestResourceAddedApplyFails(t *testing.T) {
	dir := createTestDir(testTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	mockKube.EXPECT().Apply(gomock.Any()).Return("", errors.New("apply failed"))

	ct := counterTest{
		events:    1,
		createErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { c.ResourceAdded(testResource) }, tsExpected)
}

func TestResourceAddedTemplatingFails(t *testing.T) {
	dir := createTestDir(testBadTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	ct := counterTest{
		events:    1,
		createErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { c.ResourceAdded(testResource) }, tsExpected)
}

func TestResourceDeletedHappyPath(t *testing.T) {
	dir := createTestDir(testTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	mockKube.EXPECT().Delete(gomock.Any())

	ct := counterTest{
		events:   1,
		delete:   1,
		releases: -1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_delete_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { c.ResourceDeleted(testResource) }, tsExpected)
}

func TestResourceDeletedApplyFails(t *testing.T) {
	dir := createTestDir(testTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	mockKube.EXPECT().Delete(gomock.Any()).Return("", errors.New("apply failed"))

	ct := counterTest{
		events:    1,
		deleteErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { c.ResourceDeleted(testResource) }, tsExpected)
}

func TestResourceDeletedTemplatingFails(t *testing.T) {
	dir := createTestDir(testBadTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	ct := counterTest{
		events:    1,
		deleteErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { c.ResourceDeleted(testResource) }, tsExpected)
}

func TestResourceUpdatedHappyPath(t *testing.T) {
	dir := createTestDir(testTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	mockKube.EXPECT().Apply(gomock.Any())

	ct := counterTest{
		events: 1,
		update: 1,
	}
	tsExpected := timestampTestMap()
	tsExpected["releases_last_update_timestamp_utc_seconds"] = greaterThan

	assertMetrics(t, ct, func() { c.ResourceUpdated(testResource, testResource) }, tsExpected)
}

func TestResourceUpdatedApplyFails(t *testing.T) {
	dir := createTestDir(testTemplates)
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)
	c := tmplctlr.NewController(dir, "", nil)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockKube := NewMockKubeClient(mockCtrl)
	c.Client = mockKube

	mockKube.EXPECT().Apply(gomock.Any()).Return("", errors.New("apply failed"))

	ct := counterTest{
		events:    1,
		updateErr: 1,
	}
	tsExpected := timestampTestMap()

	assertMetrics(t, ct, func() { c.ResourceUpdated(testResource, testResource) }, tsExpected)
}
