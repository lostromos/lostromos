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
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTemplate(t *testing.T) {
	// Here we create a temporary directory and populate it with our sample
	// template definition files; usually the template files would already
	// exist in some location known to the program.
	dir := createTestDir([]templateFile{
		// T0.tmpl is a plain template file that just invokes T1.
		{"0_base.tmpl", `--- {{template "file1.tmpl" . }}`},
		// T1.tmpl defines a template, T1 that invokes T2.
		{"file1.tmpl", `name: {{ .GetField "metadata" "name"  }}-configmap`},
	})
	// Clean up after the test; another quirk of running as an example.
	defer os.RemoveAll(dir)

	c := NewController(dir, "")
	tmpFile, err := c.generateTemplate(basicCR)

	assert.Nil(t, err)
	assert.NotNil(t, tmpFile)

	fmt.Printf("temp file: %s\n", tmpFile)

	bytes, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		log.Fatal("Failed to read file: " + tmpFile)
	}

	assert.Equal(t, "--- name: dory-configmap", string(bytes[:]))
}

func TestGenerateTemplateNoTemplatePath(t *testing.T) {
	c := NewController("", "")
	tmpFile, err := c.generateTemplate(basicCR)

	assert.NotNil(t, err)
	assert.Empty(t, tmpFile)
}
