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
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wpengine/lostromos/crwatcher"
	"github.com/wpengine/lostromos/printctlr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: `Start the server.`,
	Run: func(command *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	LostromosCmd.AddCommand(startCmd)
	startCmd.Flags().String("kube-master", "", "address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	startCmd.Flags().String("kube-config", filepath.Join(homeDir(), ".kube", "config"), "absolute path to the kubeconfig file. Only required if out-of-cluster.")
	startCmd.Flags().Bool("k8s", false, "When set uses the K8s service account token. Use when running in a cluter. This takes precedence over the kube-master and kube-config settings.")
	startCmd.Flags().String("crd-name", "", "the plural name of the CRD you want monitored (ex: users)")
	startCmd.Flags().String("crd-group", "", "the group of the CRD you want monitored (ex: stable.wpengine.io)")
	startCmd.Flags().String("crd-version", "v1", "the version of the CRD you want monitored")
	startCmd.Flags().String("crd-namespace", metav1.NamespaceNone, "(optional) the namespace of the CRD you want monitored, only needed for namespaced CRDs (ex: default)")

	viperBindFlag("k8s.master", startCmd.Flags().Lookup("kube-master"))
	viperBindFlag("k8s.config", startCmd.Flags().Lookup("kube-config"))
	viperBindFlag("k8s.in-cluster", startCmd.Flags().Lookup("k8s"))
	viperBindFlag("crd.name", startCmd.Flags().Lookup("crd-name"))
	viperBindFlag("crd.group", startCmd.Flags().Lookup("crd-group"))
	viperBindFlag("crd.version", startCmd.Flags().Lookup("crd-version"))
	viperBindFlag("crd.namespace", startCmd.Flags().Lookup("crd-namespace"))
}

func homeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return home
}

func getKubeClient() *restclient.Config {
	var (
		cfg *restclient.Config
		err error
	)
	if viper.GetBool("k8s.in-cluster") {
		cfg, err = restclient.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(viper.GetString("k8s.master"), viper.GetString("k8s.config"))
	}
	if err != nil {
		panic(err.Error())
	}
	return cfg
}

func buildCRWatcher(cfg *restclient.Config) *crwatcher.CRWatcher {
	cwCfg := &crwatcher.Config{
		PluralName: viper.GetString("crd.name"),
		Group:      viper.GetString("crd.group"),
		Version:    viper.GetString("crd.version"),
		Namespace:  viper.GetString("crd.namespace"),
	}

	crw, err := crwatcher.NewCRWatcher(cwCfg, cfg, printctlr.Controller{})
	if err != nil {
		panic(err.Error())
	}
	return crw
}

func startServer() {
	cfg := getKubeClient()
	crw := buildCRWatcher(cfg)
	err := crw.Watch(wait.NeverStop)
	if err != nil {
		panic(err.Error())
	}
}
