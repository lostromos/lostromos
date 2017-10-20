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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// LostromosCmd represents the base command when called without any subcommands
var LostromosCmd = &cobra.Command{
	Use:   "lostromos",
	Short: "Create K8s resources from Custom Resources",
	Long: `Lostromos will monitor all the resources created in your K8s CRD
and create, update, and delete resources based on the templates provided.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the LostromosCmd.
func Execute() {
	if err := LostromosCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	LostromosCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/lostromos.yaml)")
	LostromosCmd.PersistentFlags().BoolP("debug", "", false, "enable debug logging")

	LostromosCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Use default config file /etc/lostromos.yaml
		viper.AddConfigPath("/etc")
		viper.SetConfigName("lostromos.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
