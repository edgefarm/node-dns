/*
Copyright © 2021 Armin Schlegel <armin.schlegel@gmx.de>

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
package config

type FeedConfig struct {
	// K8sapi configures the k8s api feed
	K8sapi K8sApiConfig
}

type K8sApiConfig struct {
	// Enabled indicates if the k8s api feed is used
	// default: false
	Enabled bool `json:"enabled"`
	// URI is where the api server is reacheble. Format: 'host:port', optional with 'http://' or 'https://'
	// default: http://127.0.0.1:10550
	URI string `json:"URI"`
	// InsecureTLS indicates if there is any TLS certificate used that is self signed (optional)
	// default: true
	InsecureTLS bool `json:"insecureTLS"`
	// Token is the token to communicate with the API server (optional)
	// default: ""
	Token string `json:"token"`
}

func NewFeedConfig() *FeedConfig {
	return &FeedConfig{
		K8sapi: K8sApiConfig{
			Enabled:     true,
			URI:         "http://127.0.0.1:10550",
			InsecureTLS: true,
			Token:       "",
		},
	}
}
