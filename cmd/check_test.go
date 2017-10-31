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
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var checktests = []struct {
	tmplDir  string
	crFile   string
	out      string
	errors   bool
	errorOut string
}{
	{"/path/not/found", "../test/data/cr_nemo.yml", "", true, "ERROR: your templates directory does not exist"},
	{"../test/data/templates/0_base.tmpl", "../test/data/cr_nemo.yml", "", true, "ERROR: your templates directory is not a directory"},
	{"../test/data/templates/", "/path/not/found", "", true, "ERROR: your CR file does not exist"},
	{"../test/data/templates/", "../test/data/", "", true, "ERROR: your CR file is not a file"},
	{"../test/data/templates/", "../test/data/cr_nemo.yml", validtemplate, false, ""},
}

var validtemplate = "---\n\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: nemo-configmap\ndata:\n  by: Disney\n\n\n---\n\napiVersion: apps/v1beta1\nkind: Deployment\nmetadata:\n  name: nemo-nginx\nspec:\n  replicas: 1\n  template:\n    metadata:\n      labels:\n        app: nemo\n        component: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:alpine\n        ports:\n        - containerPort: 80\n\n"

func TestCheckCommand(t *testing.T) {
	for _, tt := range checktests {
		tmplDir = tt.tmplDir
		crFile = tt.crFile
		var b bytes.Buffer
		w := bufio.NewWriter(&b)

		err := check(w)

		w.Flush()

		assert.Equal(t, tt.out, b.String())
		if tt.errors {
			assert.NotNil(t, err)
			assert.Equal(t, tt.errorOut, err.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}
