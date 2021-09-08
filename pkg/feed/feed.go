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
package feed

// Feed is an interface to enable different sources to obtain of host/ip entries
type FeedIf interface {
	Update() error
	GetDnsMap() map[string]string
}

type Feed struct {
	// FeedDnsMap is a map of hostnames with their corresponding IP addresses
	FeedDnsMap map[string]string
}
