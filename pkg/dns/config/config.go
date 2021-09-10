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

package config

import (
	feed "github.com/siredmar/node-dns/pkg/feed/config"
)

// DNSConfig contains the configuration for the DNS
type DNSConfig struct {
	// ListenInterface defines the interface the edgeDNS listens on
	// default: docker0
	ListenInterface string `json:"listenInterface"`
	//  ListenPort defines the DNS port
	// default: 53
	ListenPort int `json:"listenPort"`
	// Feed defines the feeds the DNS server gets its information from
	Feed *feed.FeedConfig `json:"feed"`
}

// NewDNSConfig gets the default DNS configuration
func NewDNSConfig() *DNSConfig {
	return &DNSConfig{
		ListenInterface: "docker0",
		ListenPort:      53,
		Feed:            feed.NewFeedConfig(),
	}
}
