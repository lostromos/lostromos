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
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourcePhase string

const (
	PhaseNone     ResourcePhase = ""
	PhaseApplying ResourcePhase = "Applying"
	PhaseApplied  ResourcePhase = "Applied"
	PhaseFailed   ResourcePhase = "Failed"
)

type ConditionReason string

const (
	ReasonUnknown               ConditionReason = "Unknown"
	ReasonCustomResourceAdded   ConditionReason = "CustomResourceAdded"
	ReasonCustomResourceUpdated ConditionReason = "CustomResourceUpdated"
	ReasonApplySuccessful       ConditionReason = "ApplySuccessful"
	ReasonApplyFailed           ConditionReason = "ApplyFailed"
)

type CustomResourceStatus struct {
	Phase              ResourcePhase   `json:"phase"`
	Reason             ConditionReason `json:"reason,omitempty"`
	Message            string          `json:"message,omitempty"`
	LastUpdateTime     metav1.Time     `json:"lastUpdateTime,omitempty"`
	LastTransitionTime metav1.Time     `json:"lastTransitionTime,omitempty"`
}

// SetPhase takes a custom resource status and returns the updated status, without updating the resource in the cluster.
func SetPhase(status CustomResourceStatus, phase ResourcePhase, reason ConditionReason, message string) map[string]interface{} {
	status.LastUpdateTime = metav1.Now()
	if status.Phase != phase {
		status.Phase = phase
		status.LastTransitionTime = metav1.Now()
	}
	status.Message = message
	status.Reason = reason

	var out map[string]interface{}
	jsonObj, err := json.Marshal(&status)
	if err != nil {
		return map[string]interface{}{}
	}
	json.Unmarshal(jsonObj, &out)
	return out
}

// StatusFor safely returns a typed status block from a custom resource.
func StatusFor(cr *unstructured.Unstructured) CustomResourceStatus {
	switch cr.Object["status"].(type) {
	case CustomResourceStatus:
		return cr.Object["status"].(CustomResourceStatus)
	case map[string]interface{}:
		var status CustomResourceStatus
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(cr.Object["status"].(map[string]interface{}), &status); err != nil {
			return CustomResourceStatus{
				Phase:   PhaseFailed,
				Reason:  ReasonApplyFailed,
				Message: err.Error(),
			}
		}
		return status
	default:
		return CustomResourceStatus{}
	}
}
