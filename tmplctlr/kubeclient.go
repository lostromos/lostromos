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

package tmplctlr

import (
	"os"
	"os/exec"
)

// KubeClient is an interface that implements an Apply() and Delete() for our K8s templates
type KubeClient interface {
	Apply(file string) (string, error)
	Delete(file string) (string, error)
}

// Kubectl provides a simple wrapper around calling the needed kubectl commands
// TODO: This should be revisited when https://github.com/kubernetes/kubernetes/issues/15894 is completed.
// #15894 will move the apply logic from kubectl into the API
type Kubectl struct {
	ConfigFile string //config file for kubectl
}

var execCommand = exec.Command

// Apply will execute kubectl apply -f file with the correct config
func (k Kubectl) Apply(file string) (string, error) {
	if k.ConfigFile != "" {
		err := os.Setenv("KUBECONFIG", k.ConfigFile)
		if err != nil {
			return "", err
		}
	}
	out, err := execCommand("kubectl", "apply", "-f", file).CombinedOutput()
	return string(out[:]), err
}

// Delete will execute kubectl delete -f file with the correct config
func (k Kubectl) Delete(file string) (string, error) {
	if k.ConfigFile != "" {
		err := os.Setenv("KUBECONFIG", k.ConfigFile)
		if err != nil {
			return "", err
		}
	}
	out, err := execCommand("kubectl", "delete", "-f", file).CombinedOutput()
	return string(out[:]), err
}
