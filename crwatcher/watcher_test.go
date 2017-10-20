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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wpengine/lostromos/printctlr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
)

func TestNewCRWatcher(t *testing.T) {
	kubeCfg := &restclient.Config{}
	cfg := &Config{PluralName: "test"}

	cw, err := NewCRWatcher(cfg, kubeCfg, printctlr.Controller{})

	assert.Nil(t, err)
	assert.Equal(t, cfg, cw.Config)
	assert.NotNil(t, cw.resource)
	assert.NotNil(t, cw.handler)
	assert.NotNil(t, cw.store)
	assert.NotNil(t, cw.controller)
}

func TestSetupHandlerAddFunc(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{}
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

func TestSetupHandlerDeleteFunc(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{}
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

func TestSetupHandlerUpdateFunc(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRC := NewMockResourceController(mockCtrl)
	cw := &CRWatcher{}
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

func TestWatchReturnsErrorIfNotSetup(t *testing.T) {
	cw := &CRWatcher{}
	err := cw.Watch(wait.NeverStop)

	assert.NotNil(t, err)
}
