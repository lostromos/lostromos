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

package crwatcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	cr "github.com/wpengine/lostromos/crwatcher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	testNamespaceName = "lostromos-test"
	testReleaseName   = "lostromostest-dory"
)

var now = metav1.Now()

func TestSetPhase(t *testing.T) {
	newStatus := cr.SetPhase(newTestStatus(), cr.PhaseApplying, cr.ReasonCustomResourceUpdated, "working on it")

	assert.Equal(t, string(cr.PhaseApplying), newStatus["phase"])
	assert.Equal(t, string(cr.ReasonCustomResourceUpdated), newStatus["reason"])
	assert.Equal(t, "working on it", newStatus["message"])
	assert.NotEqual(t, metav1.Now(), newStatus["lastUpdateTime"])
	assert.NotEqual(t, metav1.Now(), newStatus["lastTransitionTime"])
}

func TestStatusForEmpty(t *testing.T) {
	status := cr.StatusFor(newTestResource())

	assert.Equal(t, cr.CustomResourceStatus{}, status)
}

func TestStatusForFilled(t *testing.T) {
	expectedResource := newTestResource()
	expectedResource.Object["status"] = newTestStatusRaw()
	status := cr.StatusFor(expectedResource)

	assert.EqualValues(t, newTestStatus().Phase, status.Phase)
	assert.EqualValues(t, newTestStatus().Reason, status.Reason)
	assert.EqualValues(t, newTestStatus().Message, status.Message)
}

func newTestResource() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "Character",
			"apiVersion": "stable.nicolerenee.io",
			"metadata": map[string]interface{}{
				"name":      "dory",
				"namespace": testNamespaceName,
			},
			"spec": map[string]interface{}{
				"Name": "Dory",
				"From": "Finding Nemo",
				"By":   "Disney",
			},
		},
	}
}

func newTestStatus() cr.CustomResourceStatus {
	return cr.CustomResourceStatus{
		Phase:              cr.PhaseApplied,
		Reason:             cr.ReasonApplySuccessful,
		Message:            "some message",
		LastUpdateTime:     now,
		LastTransitionTime: now,
	}
}

func newTestStatusRaw() map[string]interface{} {
	return map[string]interface{}{
		"phase":              cr.PhaseApplied,
		"reason":             cr.ReasonApplySuccessful,
		"message":            "some message",
		"lastUpdateTime":     now.UTC(),
		"lastTransitionTime": now.UTC(),
	}
}
