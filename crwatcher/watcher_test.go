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
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wpengine/lostromos/printctlr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
)

type logResult struct {
	msg string
}

type testLogger struct {
	res *logResult
}

func (c testLogger) Error(err error) {
	c.res.msg = fmt.Sprintf("error: %s", err)
}

func TestNewCRWatcher(t *testing.T) {
	kubeCfg := &restclient.Config{}
	kubeCfg.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   "test.lostromos.k8s",
		Version: "v1alpha1",
	}
	cfg := &Config{PluralName: "test"}

	dynClient, err := dynamic.NewClient(kubeCfg)
	require.NoError(t, err)

	cw, err := NewCRWatcher(cfg, dynClient, printctlr.Controller{}, testLogger{})
	require.NoError(t, err)

	assert.Equal(t, cfg, cw.Config)
	assert.NotNil(t, cw.resource)
	assert.NotNil(t, cw.handler)
	assert.NotNil(t, cw.store)
	assert.NotNil(t, cw.controller)
	assert.NotNil(t, cw.logger)
}

func TestLogKubeError(t *testing.T) {
	kubeCfg := &restclient.Config{}
	kubeCfg.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   "test.lostromos.k8s",
		Version: "v1alpha1",
	}
	cfg := &Config{PluralName: "test"}
	res := &logResult{}
	lgr := &testLogger{res: res}

	dynClient, err := dynamic.NewClient(kubeCfg)
	require.NoError(t, err)

	cw, err := NewCRWatcher(cfg, dynClient, printctlr.Controller{}, lgr)
	require.NoError(t, err)
	assert.NotNil(t, cw.logger)

	cw.logKubeError(errors.New("test"))
	assert.Equal(t, "error: test", lgr.res.msg)
}

func TestSetupHandlerAddFunc(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{
		Config: &Config{},
	}
	r := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}
	cw.setupHandler(mockRC)

	mockRC.EXPECT().ResourceAdded(r)

	cw.handler.OnAdd(r)
}

// Test to ensure that if we are given filter criteria we only call ResourceAdded for a resource with the specified
// filter.
func TestSetupHandlerAddFuncUsesFilter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{
		Config: &Config{
			Filter: "com.wpengine.lostromos.filter",
		},
	}
	r1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}
	r2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing2",
				"annotations": map[string]interface{}{
					"com.wpengine.lostromos.filter": "true",
				},
			},
		},
	}
	cw.setupHandler(mockRC)

	mockRC.EXPECT().ResourceAdded(r1).MinTimes(0).MaxTimes(0)
	mockRC.EXPECT().ResourceAdded(r2)

	cw.handler.OnAdd(r1)
	cw.handler.OnAdd(r2)
}

func TestSetupHandlerDeleteFunc(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{
		Config: &Config{},
	}
	r := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}
	cw.setupHandler(mockRC)

	mockRC.EXPECT().ResourceDeleted(r)

	cw.handler.OnDelete(r)
}

// Test to ensure that if we are given filter criteria we only call ResourceDeleted for a resource with the specified
// filter.
func TestSetupHandlerDeleteFuncUsesFilter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{
		Config: &Config{
			Filter: "com.wpengine.lostromos.filter",
		},
	}
	r1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}
	r2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing2",
				"annotations": map[string]interface{}{
					"com.wpengine.lostromos.filter": "true",
				},
			},
		},
	}
	cw.setupHandler(mockRC)

	mockRC.EXPECT().ResourceDeleted(r1).MinTimes(0).MaxTimes(0)
	mockRC.EXPECT().ResourceDeleted(r2)

	cw.handler.OnDelete(r1)
	cw.handler.OnDelete(r2)
}

func TestSetupHandlerUpdateFunc(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{
		Config: &Config{},
	}
	r1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}
	r2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing2",
			},
		},
	}
	cw.setupHandler(mockRC)

	mockRC.EXPECT().ResourceUpdated(r1, r2)

	cw.handler.OnUpdate(r1, r2)
}

// Test to ensure that if we are given filter criteria we call ResourceUpdated, ResourceDeleted, and ResourceAdded
// appropriately for filtered and unfiltered resources.
func TestSetupHandlerUpdateFuncUsesFilter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{
		Config: &Config{
			Filter: "com.wpengine.lostromos.filter",
		},
	}
	r1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
			},
		},
	}
	r1Filtered := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing1",
				"annotations": map[string]interface{}{
					"com.wpengine.lostromos.filter": "true",
				},
			},
		},
	}
	r2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing2",
			},
		},
	}
	r2Filtered := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "Thing2",
				"annotations": map[string]interface{}{
					"com.wpengine.lostromos.filter": "true",
				},
			},
		},
	}
	cw.setupHandler(mockRC)

	mockRC.EXPECT().ResourceUpdated(r1, r2).MinTimes(0).MaxTimes(0)
	mockRC.EXPECT().ResourceDeleted(r1Filtered)
	mockRC.EXPECT().ResourceAdded(r2Filtered)
	mockRC.EXPECT().ResourceUpdated(r1Filtered, r2Filtered)

	cw.handler.OnUpdate(r1, r2)
	cw.handler.OnUpdate(r1Filtered, r2)
	cw.handler.OnUpdate(r1, r2Filtered)
	cw.handler.OnUpdate(r1Filtered, r2Filtered)
}

func TestWatchReturnsErrorIfNotSetup(t *testing.T) {
	cw := &CRWatcher{}
	err := cw.Watch(wait.NeverStop)

	assert.NotNil(t, err)
}
