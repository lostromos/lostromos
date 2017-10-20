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

package cmd

import (
	"path"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	restclient "k8s.io/client-go/rest"
)

func TestGetKubeClientForK8sPanicsWhenNotInCluster(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	viper.Set("k8s.in-cluster", true)
	cfg := getKubeClient()
	assert.Nil(t, cfg)
}

func TestGetKubeClientForKubeConfigFailsWhenFileNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	viper.Set("k8s.in-cluster", false)
	viper.Set("k8s.config", "/i-dont-exist/kubeconf")
	cfg := getKubeClient()
	assert.Nil(t, cfg)
}

func TestGetKubeClientForKubeConfigReturnsWhenFileExists(t *testing.T) {
	viper.Set("k8s.in-cluster", false)
	viper.Set("k8s.config", path.Join("..", "test-data", "kubeconfig"))
	cfg := getKubeClient()
	assert.NotNil(t, cfg)
	// This value is from the test-data/kubeconfig file
	assert.Equal(t, "https://localhost:8443", cfg.Host)
}

func TestBuildCRWatcherReturnsProperlyConfiguredWatcher(t *testing.T) {
	crdGroup := "test.lostromos.k8s"
	crdName := "testCRD"
	crdNamespace := "lostromos"
	crdVersion := "v9876"
	viper.Set("crd.group", crdGroup)
	viper.Set("crd.name", crdName)
	viper.Set("crd.namespace", crdNamespace)
	viper.Set("crd.version", crdVersion)

	kubeCfg := &restclient.Config{}
	crw := buildCRWatcher(kubeCfg)
	assert.NotNil(t, crw)
	assert.Equal(t, crdGroup, crw.Config.Group)
	assert.Equal(t, crdName, crw.Config.PluralName)
	assert.Equal(t, crdNamespace, crw.Config.Namespace)
	assert.Equal(t, crdVersion, crw.Config.Version)
}
