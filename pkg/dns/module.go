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

package dns

import (
	"fmt"
	"net"

	mdns "github.com/miekg/dns"
	"github.com/siredmar/node-dns/pkg/dns/config"

	"github.com/siredmar/node-dns/pkg/feed"
)

// EdgeDNS is a node-level dns resolver
type EdgeDNS struct {
	ListenIP net.IP
	Server   *mdns.Server
	Exit     chan interface{}
	Feed     feed.If
}

// NewEdgeDNS creates a new EdgeDNS instance
func NewEdgeDNS(config *config.DNSConfig) (dns *EdgeDNS, err error) {
	dns = &EdgeDNS{
		ListenIP: []byte{},
		Server:   &mdns.Server{},
		Exit:     make(chan interface{}),
		Feed:     feed.NewK8sAPI(config.Feed),
	}

	// get dns listen ip
	dns.ListenIP, err = getInterfaceIP(config.ListenInterface)
	if err != nil {
		return dns, fmt.Errorf("get dns listen ip for interface %s err: %v", config.ListenInterface, err)
	}

	addr := fmt.Sprintf("%v:%v", dns.ListenIP, config.ListenPort)
	dns.Server = &mdns.Server{Addr: addr, Net: "udp"}

	return dns, nil
}

// getInterfaceIP get net interface IPv4 address
func getInterfaceIP(name string) (net.IP, error) {
	ifi, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}
	addrs, _ := ifi.Addrs()
	for _, addr := range addrs {
		if ip, ipn, _ := net.ParseCIDR(addr.String()); len(ipn.Mask) == 4 {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("no ip of version 4 found for interface %s", name)
}
