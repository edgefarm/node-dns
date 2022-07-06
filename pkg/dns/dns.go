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

var (
	// DNSMap is a map of hostnames with their corresponding IP addresses
	DNSMap           = map[string]string{}
	otherNameservers = []string{}
)

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
	go func() {
		if dns.UpdateResolvConf {
			dns.ensureResolvForHost()
			otherNameservers = dns.otherNameservers()
			err := dns.ensureRemovedSearchDomains()
			if err != nil {
				klog.Errorf("%v", err)
			}
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
				klog.Infof("Currently resolvable:")
				for host, ip := range DNSMap {
					klog.Infof("  %s -> %s", host, ip)
				}
				if dns.UpdateResolvConf {
					dns.ensureResolvForHost()
					otherNameservers = dns.otherNameservers()
					err := dns.ensureRemovedSearchDomains()
					if err != nil {
						klog.Errorf("%v", err)
					}
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

func (dns *EdgeDNS) ensureRemovedSearchDomains() error {
	resolv, err := readFile(dns.ResolvConf)
	if err != nil {
		return err
	}
	resolv = dns.removeSearchDomains(resolv)
	str := strings.Join(resolv, "\n")
	if err := writeFile(dns.ResolvConf, []byte(str)); err != nil {
		return err
	}
	return nil
}

func (dns *EdgeDNS) removeSearchDomains(resolv []string) []string {
	if dns.RemoveSearchDomains {
		for i, line := range resolv {
			if strings.Contains(line, "search") {
				resolv = append(resolv[:i], resolv[i+1:]...)
			}
		}
	}
	return resolv
}

func readFile(file string) ([]string, error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file %s err: %v", file, err)
	}

	resolv := strings.Split(string(bs), "\n")
	return resolv, nil
}

func writeFile(file string, content []byte) error {
	if err := ioutil.WriteFile(file, content, 0600); err != nil {
		return fmt.Errorf("failed to write file %s, err: %v", file, err)
	}
	return nil
}

// ensureResolvForHost adds edgemesh dns server to the head of /etc/resolv.conf
func (dns *EdgeDNS) ensureResolvForHost() {
	if dns.ListenIP != nil {
		resolv, err := readFile(dns.ResolvConf)
		if err != nil {
			klog.Errorf("%v", err)
			return
		}

		if resolv == nil {
			nameserver := "nameserver " + dns.ListenIP.String()
			if err := writeFile(dns.ResolvConf, []byte(nameserver)); err != nil {
				klog.Errorf("err: %v", err)
				return
			}
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
				if err := writeFile(dns.ResolvConf, []byte(nameserver)); err != nil {
					klog.Errorf("err: %v", err)
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

		if err := writeFile(dns.ResolvConf, []byte(nameserver)); err != nil {
			klog.Errorf("err: %v", err)
			return
		}
	}
}

// otherNameservers returns a list of other nameservers configured in /etc/resolv.conf other than ours
func (dns *EdgeDNS) otherNameservers() []string {
	bs, err := ioutil.ReadFile(dns.ResolvConf)
	if err != nil {
		klog.Errorf("failed to read file %s, err: %v", dns.ResolvConf, err)
	}

	resolv := strings.Split(string(bs), "\n")

	nameservers := []string{}
	for _, line := range resolv {
		if strings.Contains(line, "nameserver") {
			ip := strings.Split(line, "nameserver ")
			nameservers = append(nameservers, ip[1])
		}

	}
	others := []string{}
	for _, ip := range nameservers {
		if ip != dns.ListenIP.String() {
			others = append(others, ip)
		}

	}
	return others
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
	bs, err := ioutil.ReadFile(dns.ResolvConf)
	if err != nil {
		klog.Warningf("read file %s err: %v", dns.ResolvConf, err)
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
	if err := ioutil.WriteFile(dns.ResolvConf, []byte(nameserver), 0600); err != nil {
		klog.Errorf("failed to write nameserver to file %s, err: %v", dns.ResolvConf, err)
	}
}

func lookupUpstreamHost(ctx context.Context, URI string) ([]string, error) {
	address := []string{}
	var lastErr error = nil
	for _, other := range otherNameservers {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(10000),
				}
				return d.DialContext(ctx, network, fmt.Sprintf("%s:53", other))
			},
		}
		found, err := r.LookupHost(context.Background(), URI)
		if err != nil {
			klog.Infof("cannot resolve %s using %s, err: %v", URI, other, err)
			lastErr = err
			continue
		}
		if len(found) > 0 && err == nil {
			address = found
			break
		}
	}
	return address, lastErr
}
