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
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Print(os.Args[3:])
	if os.Args[len(os.Args)-1] == "ERROR" {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestKubectlApply(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	k := &Kubectl{}
	out, err := k.Apply("path")
	assert.Nil(t, err)
	assert.Equal(t, "[kubectl apply -f path]", out)
}

func TestKubectlApplyCmdError(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	k := &Kubectl{}
	out, err := k.Apply("ERROR")
	assert.NotNil(t, err)
	assert.Equal(t, "[kubectl apply -f ERROR]", out)
}

func TestKubectlApplyConfigFile(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	k := &Kubectl{ConfigFile: "some_file"}
	out, err := k.Apply("path")
	assert.Nil(t, err)
	assert.Equal(t, "some_file", os.Getenv("KUBECONFIG"))
	assert.Equal(t, "[kubectl apply -f path]", out)
}

func TestKubectlDelete(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	k := &Kubectl{}
	out, err := k.Delete("path")
	assert.Nil(t, err)
	assert.Equal(t, "[kubectl delete -f path]", out)
}

func TestKubectlDeleteConfigFile(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	k := &Kubectl{ConfigFile: "some_file"}
	out, err := k.Delete("path")
	assert.Nil(t, err)
	assert.Equal(t, "some_file", os.Getenv("KUBECONFIG"))
	assert.Equal(t, "[kubectl delete -f path]", out)
}

func TestKubectlDeleteCmdError(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	k := &Kubectl{}
	out, err := k.Delete("ERROR")
	assert.NotNil(t, err)
	assert.Equal(t, "[kubectl delete -f ERROR]", out)
}
