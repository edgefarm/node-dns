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

package feed

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/siredmar/node-dns/pkg/feed/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	podsAPI = "/api/v1/pods"
)

// K8sAPI defines the k8s api feed
type K8sAPI struct {
	Feed
	URI         string
	Token       string
	InsecureTLS bool
}

// NewK8sAPI creates a new feed using the k8s API
func NewK8sAPI(config *config.FeedConfig) *K8sAPI {
	klog.Info("Starting local k8s api feed")
	return &K8sAPI{
		URI:         config.K8sapi.URI,
		Token:       config.K8sapi.Token,
		InsecureTLS: config.K8sapi.InsecureTLS,
		Feed: Feed{
			FeedDNSMap: make(map[string]string),
		},
	}
}

// Update triggers an update of the DNS cache
func (k8s *K8sAPI) Update() error {
	klog.Info("Updating DNS cache")
	podsRaw, err := k8s.getPods()
	if err != nil {
		return err
	}
	podIPs, err := k8s.getPodIPs(podsRaw)
	if err != nil {
		return err
	}

	for k := range k8s.Feed.FeedDNSMap {
		delete(k8s.Feed.FeedDNSMap, k)
	}
	for host, ip := range podIPs {
		k8s.Feed.FeedDNSMap[host] = ip
	}

	return nil
}

// GetDNSMap returns the feeds DNS map
func (k8s *K8sAPI) GetDNSMap() map[string]string {
	return k8s.Feed.FeedDNSMap
}

// getPodIPs extracts the IPs from the pods
func (k8s *K8sAPI) getPodIPs(podlist *corev1.PodList) (map[string]string, error) {
	podIPs := map[string]string{}
	for _, pod := range podlist.Items {
		if podName, ok := pod.Labels["node-dns.host"]; ok {
			for _, container := range pod.Spec.Containers {
				podIPs[fmt.Sprintf("%s.%s", container.Name, podName)] = pod.Status.PodIP
			}
		}
	}
	return podIPs, nil
}

// getPods gets all pods from the k8s api
func (k8s *K8sAPI) getPods() (*corev1.PodList, error) {

	// Create a new request using http
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", k8s.URI, podsAPI), nil)
	if err != nil {
		return nil, err
	}

	if len(k8s.Token) > 0 {
		// Create a Bearer string by appending string access token
		var bearer = "Bearer " + k8s.Token
		// add authorization header to the req
		req.Header.Add("Authorization", bearer)
	}

	// Send req using http Client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: k8s.InsecureTLS},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	podlist := &corev1.PodList{}

	err = json.Unmarshal(body, &podlist)
	if err != nil {
		return nil, err
	}
	return podlist, nil
}
