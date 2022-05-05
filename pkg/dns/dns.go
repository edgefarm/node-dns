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
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	mdns "github.com/miekg/dns"
	"k8s.io/klog/v2"
)

const hostResolv = "/etc/resolv.conf"

const (
	customDNS = "8.8.8.8:53"
)

// DNSMap is a map of hostnames with their corresponding IP addresses
var DNSMap = map[string]string{}

type handler struct{}

// ServeDNS handles the DNS requests
func (h *handler) ServeDNS(w mdns.ResponseWriter, r *mdns.Msg) {
	msg := mdns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case mdns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		domainTrimmed := strings.TrimRight(domain, ".")
		address, ok := lookup(domainTrimmed)
		if ok {
			msg.Answer = append(msg.Answer, &mdns.A{
				Hdr: mdns.RR_Header{Name: domain, Rrtype: mdns.TypeA, Class: mdns.ClassINET, Ttl: 60},
				A:   address,
			})
		} else {
			return
		}
	}
	if err := w.WriteMsg(&msg); err != nil {
		klog.Errorf("dns response send error: %v", err)
	}
}

// Run starts the DNS server
func (dns *EdgeDNS) Run() {
	// ensure /etc/resolv.conf have dns nameserver
	go func() {
		if dns.UpdateResolvConf {
			dns.ensureResolvForHost()
		}
		err := dns.Feed.Update()
		if err != nil {
			klog.Errorf("failed to update dns server, err: %v", err)
		}
		DNSMap = dns.Feed.GetDNSMap()
		klog.Infof("Currently resolvable:")
		for host, ip := range DNSMap {
			klog.Infof("  %s -> %s", host, ip)
		}
		ticker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-ticker.C:
				err := dns.Feed.Update()
				if err != nil {
					klog.Errorf("failed to update dns server, err: %v", err)
				}
				DNSMap = dns.Feed.GetDNSMap()
				if dns.UpdateResolvConf {
					fmt.Println(DNSMap)
					dns.ensureResolvForHost()
				}
			case <-dns.Exit:
				if dns.UpdateResolvConf {
					dns.cleanResolvForHost()
				}
				return
			}
		}
	}()
	dns.Server.Handler = &handler{}
	if err := dns.Server.ListenAndServe(); err != nil {
		klog.Errorf("dns server serve error: %v", err)
	}
}

// Stop stops the DNS server
func (dns *EdgeDNS) Stop() error {
	dns.Exit <- true
	err := dns.Server.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

// getIPForURI returns the IP for an URI
func getIPForURI(URI string) (string, error) {
	if ip, ok := DNSMap[URI]; ok {
		return ip, nil
	}
	ips, err := lookupUpstreamHost(context.Background(), URI)
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no IP found for %s", URI)
	}
	return ips[0], nil
}

// lookup confirms if the service exists
func lookup(URI string) (ip net.IP, exist bool) {
	ipAddress, err := getIPForURI(URI)
	if err != nil {
		klog.Warningf("%v", err)
		return nil, false
	}
	klog.Infof("dns server parse %s ip %s", URI, ipAddress)
	return net.ParseIP(ipAddress), true
}

// ensureResolvForHost adds edgemesh dns server to the head of /etc/resolv.conf
func (dns *EdgeDNS) ensureResolvForHost() {
	bs, err := ioutil.ReadFile(hostResolv)
	if err != nil {
		klog.Errorf("read file %s err: %v", hostResolv, err)
		return
	}

	resolv := strings.Split(string(bs), "\n")
	if resolv == nil {
		nameserver := "nameserver " + dns.ListenIP.String()
		if err := ioutil.WriteFile(hostResolv, []byte(nameserver), 0600); err != nil {
			klog.Errorf("write file %s err: %v", hostResolv, err)
		}
		return
	}

	configured := false
	dnsIdx := 0
	startIdx := 0
	for idx, item := range resolv {
		if strings.Contains(item, dns.ListenIP.String()) {
			configured = true
			dnsIdx = idx
			break
		}
	}
	for idx, item := range resolv {
		if strings.Contains(item, "nameserver") {
			startIdx = idx
			break
		}
	}
	if configured {
		if dnsIdx != startIdx && dnsIdx > startIdx {
			nameserver := sortNameserver(resolv, dnsIdx, startIdx)
			if err := ioutil.WriteFile(hostResolv, []byte(nameserver), 0600); err != nil {
				klog.Errorf("failed to write file %s, err: %v", hostResolv, err)
				return
			}
		}
		return
	}

	nameserver := ""
	for idx := 0; idx < len(resolv); {
		if idx == startIdx {
			startIdx = -1
			nameserver = nameserver + "nameserver " + dns.ListenIP.String() + "\n"
			continue
		}
		nameserver = nameserver + resolv[idx] + "\n"
		idx++
	}

	if err := ioutil.WriteFile(hostResolv, []byte(nameserver), 0600); err != nil {
		klog.Errorf("failed to write file %s, err: %v", hostResolv, err)
		return
	}
}

// sortNameserver sorts the nameserver list
func sortNameserver(resolv []string, dnsIdx, startIdx int) string {
	nameserver := ""
	idx := 0
	for ; idx < startIdx; idx++ {
		nameserver = nameserver + resolv[idx] + "\n"
	}
	nameserver = nameserver + resolv[dnsIdx] + "\n"

	for idx = startIdx; idx < len(resolv); idx++ {
		if idx == dnsIdx {
			continue
		}
		nameserver = nameserver + resolv[idx] + "\n"
	}

	return nameserver
}

// cleanResolvForHost delete edgemesh dns server from the head of /etc/resolv.conf
func (dns *EdgeDNS) cleanResolvForHost() {
	bs, err := ioutil.ReadFile(hostResolv)
	if err != nil {
		klog.Warningf("read file %s err: %v", hostResolv, err)
	}

	resolv := strings.Split(string(bs), "\n")
	if resolv == nil {
		return
	}
	nameserver := ""
	for _, item := range resolv {
		if strings.Contains(item, dns.ListenIP.String()) || item == "" {
			continue
		}
		nameserver = nameserver + item + "\n"
	}
	if err := ioutil.WriteFile(hostResolv, []byte(nameserver), 0600); err != nil {
		klog.Errorf("failed to write nameserver to file %s, err: %v", hostResolv, err)
	}
}

func lookupUpstreamHost(ctx context.Context, URI string) ([]string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, customDNS)
		},
	}
	return r.LookupHost(context.Background(), URI)
}
