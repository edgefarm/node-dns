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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	dns "github.com/siredmar/node-dns/pkg/dns"
	"github.com/siredmar/node-dns/pkg/dns/config"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "node-dns",
	Short: "node-dns is a simple dns server for a k8s node.",
	Long: `node-dns is a simple dns server for a k8s node.

If your POD is annotated with 'node-dns.host: <pod>', it
can be resolved using this server. node-dns patches the
'/etc/resolv.conf' to make it available to, e.g. docker
containers running on the host.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		config := config.NewDNSConfig()
		config.ListenInterface = viper.GetString("listeninterface")
		config.ListenPort = viper.GetInt("listenport")
		config.Feed.K8sapi.Enabled = viper.GetBool("feed.k8sapi.enabled")
		config.Feed.K8sapi.InsecureTLS = viper.GetBool("feed.k8sapi.insecuretls")
		config.Feed.K8sapi.Token = viper.GetString("feed.k8sapi.token")
		config.Feed.K8sapi.URI = viper.GetString("feed.k8sapi.uri")
		dns, err := dns.NewEdgeDNS(config)
		if err != nil {
			klog.Errorf("Error creating DNS: %v", err)
			os.Exit(1)
		}
		klog.Infof("Starting DNS server")
		dns.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.node-dns.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".node-dns" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".node-dns")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		klog.Errorf("Failed reading config file: %s", err)
		os.Exit(1)
	}
}
