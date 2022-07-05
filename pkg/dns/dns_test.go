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
	"log"
	"os"
	"testing"

	"github.com/edgefarm/node-dns/pkg/dns/config"
	"github.com/stretchr/testify/assert"
)

var (
	predefinedResolvConf string = `
search svc.cluster.local cluster.local
nameserver 8.8.8.8
nameserver 4.4.4.4
`
)

func setupEdgeDNS(t *testing.T) (*EdgeDNS, string) {
	assert := assert.New(t)
	file, err := os.CreateTemp("", "resolvconf")
	if err != nil {
		log.Fatal(err)
	}

	config := config.NewDNSConfig()
	config.ResolvConf = file.Name()

	_, err = file.WriteString(predefinedResolvConf)
	assert.Nil(err)

	e, err := NewEdgeDNS(config)
	assert.Nil(err)
	return e, file.Name()
}

func cleanupEdgeDNS(t *testing.T, file string) {
	os.Remove(file)
}

func TestLookupUptreamHost(t *testing.T) {
	assert := assert.New(t)
	_, file := setupEdgeDNS(t)
	defer cleanupEdgeDNS(t, file)
	ips, err := lookupUpstreamHost(context.Background(), "example.com")
	assert.Nil(err)
	assert.NotEmpty(ips)

	ips, err = lookupUpstreamHost(context.Background(), "foo-bar-this-is-never-found.com")
	fmt.Println(err)
	assert.NotNil(err)
	assert.Empty(ips)
}

func TestGetOtherNameservers(t *testing.T) {
	assert := assert.New(t)
	e, file := setupEdgeDNS(t)
	defer cleanupEdgeDNS(t, file)
	others := e.otherNameservers()
	assert.Contains(others, "8.8.8.8")
	assert.Contains(others, "4.4.4.4")
	assert.Equal(len(others), 2)
}
