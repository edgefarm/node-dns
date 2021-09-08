package config

import (
	feed "github.com/siredmar/node-dns/pkg/feed/config"
)

type DnsConfig struct {
	// ListenInterface defines the interface the edgeDNS listens on
	// default: docker0
	ListenInterface string `json:"listenInterface"`
	//  ListenPort defines the DNS port
	// default: 53
	ListenPort int `json:"listenPort"`
	// Feed defines the feeds the DNS server gets its information from
	Feed *feed.FeedConfig `json:"feed"`
}

func NewDnsConfig() *DnsConfig {
	return &DnsConfig{
		ListenInterface: "docker0",
		ListenPort:      53,
		Feed:            feed.NewFeedConfig(),
	}
}
