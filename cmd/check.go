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
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/lostromos/lostromos/tmpl"
)

var (
	crFile  string
	tmplDir string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: `Check that results of your template using the given CR.`,
	Run: func(command *cobra.Command, args []string) {
		err := check(os.Stdout)
		if err != nil {
			logger.Errorw("failed", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	LostromosCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&crFile, "cr", "", "absolute path to a yaml file with your CR saved in it")
	checkCmd.Flags().StringVar(&tmplDir, "templates", "", "absolute path to the directory with your template files")
}

func check(out io.Writer) error {
	src, err := os.Stat(tmplDir)
	if os.IsNotExist(err) {
		return errors.New("ERROR: your templates directory does not exist")
	}
	if !src.IsDir() {
		return errors.New("ERROR: your templates directory is not a directory")
	}

	src, err = os.Stat(crFile)
	if os.IsNotExist(err) {
		return errors.New("ERROR: your CR file does not exist")
	}
	if !src.Mode().IsRegular() {
		return errors.New("ERROR: your CR file is not a file")
	}

	yamlFile, err := ioutil.ReadFile(crFile) // nolint: gosec
	if err != nil {
		return err
	}

	var r unstructured.Unstructured

	json, err := yaml.YAMLToJSON(yamlFile)
	if err != nil {
		return err
	}

	err = r.UnmarshalJSON(json)
	if err != nil {
		return err
	}
	cr := &tmpl.CustomResource{Resource: &r}
	return tmpl.Parse(cr, filepath.Join(tmplDir, "*.tmpl"), out)
}
