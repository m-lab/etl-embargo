/*
Copyright 2013 Google Inc.
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

package embargo_test

import (
	"testing"

	embargo "github.com/m-lab/etl-embargo"
)

func TestReadWhitelistFromLocal(t *testing.T) {
	embargo_check := new(embargo.EmbargoCheck)
	embargo_check.ReadWhitelistFromLocal("testdata/whitelist")
	if !embargo_check.Whitelist["213.244.128.170"] {
		t.Error("missing IP in Whitelist: want '213.244.128.170'\n")
	}
	if embargo_check.Whitelist["2001:4c08:2003:2::16"] {
		t.Error("IP 2001:4c08:2003:2::16 should not be in Whitelist.\n")
	}
	return
}

func TestParseJson(t *testing.T) {
	body := []byte(`[
  {
    "hostname": "mlab2.acc02.measurement-lab.org", 
    "ipv4": "196.49.14.214", 
    "ipv6": ""
  }, 
  {
    "hostname": "mlab3.acc02.measurement-lab.org", 
    "ipv4": "196.49.14.227", 
    "ipv6": ""
  }, 
  {
    "hostname": "mlab1.acc02.measurement-lab.org", 
    "ipv4": "196.49.14.201", 
    "ipv6": ""
  }
]`)
	whitelist, err := embargo.ParseJson(body)
	if err != nil || len(whitelist) != 3 || !whitelist["196.49.14.227"] {
		t.Error("Do not parse bytes into struct correctly.")
	}
}
