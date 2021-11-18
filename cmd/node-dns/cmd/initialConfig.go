/*
Copyright © 2021 Ci4Rail GmbH <engineering@ci4rail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bytes"
	"log"

	"github.com/edgefarm/node-dns/pkg/dns/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"k8s.io/klog"
)

var (
	outFile string
)

// initialConfigCmd represents the initialConfig command
var initialConfigCmd = &cobra.Command{
	Use:   "initialConfig",
	Short: "Writes a basic initial configuration",
	Long:  "Writes a basic initial configuration",
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigType("yaml")
		defaultConfig := config.NewDNSConfig()
		out, err := yaml.Marshal(defaultConfig)
		if err != nil {
			log.Fatal(err)
		}
		err = viper.ReadConfig(bytes.NewBuffer(out))
		if err != nil {
			log.Fatal(err)
		}
		err = writeInitialConfig()
		if err != nil {
			log.Fatal(err)
		}
		klog.Infof("Written inital config to %s", viper.ConfigFileUsed())
	},
}

func init() {
	rootCmd.AddCommand(initialConfigCmd)

	rootCmd.PersistentFlags().StringVar(&outFile, "out", "", "output file")
}

func writeInitialConfig() error {
	if outFile != "" {
		viper.SetConfigFile(outFile)
	}
	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
