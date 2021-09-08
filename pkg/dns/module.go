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
	Feed     feed.FeedIf
}

// NewEdgeDNS creates a new EdgeDNS instance
func NewEdgeDNS(config *config.DnsConfig) (dns *EdgeDNS, err error) {
	dns = &EdgeDNS{
		ListenIP: []byte{},
		Server:   &mdns.Server{},
		Exit:     make(chan interface{}),
		Feed:     feed.NewK8sApi(config.Feed),
	}

	// get dns listen ip
	dns.ListenIP, err = getInterfaceIP(config.ListenInterface)
	if err != nil {
		return dns, fmt.Errorf("get dns listen ip err: %v", err)
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
