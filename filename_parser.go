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

// Parse filename and return componants like log-time, IP, etc.
// Filename example: 20170315T01:00:00Z_173.205.3.39_0.web100
package embargo

import (
	"strings"

	"github.com/m-lab/etl-embargo/metrics"
	"github.com/m-lab/etl/web100"
)

type FileName struct {
	Name string
}

// GetLocalIP parse the filename and return IP. For old format, it will return empty string.
func (f *FileName) GetLocalIP() string {
	localIPStart := strings.IndexByte(f.Name, '_')
	localIPEnd := strings.LastIndexByte(f.Name, '_')
	if localIPStart < 0 || localIPEnd < 0 || localIPStart >= localIPEnd {
		return ""
	}
	ip, err := web100.NormalizeIPv6(f.Name[localIPStart+1 : localIPEnd])
	if err != nil {
		metrics.IPv6ErrorsTotal.WithLabelValues(err.Error()).Inc()
		return ""
	}
	return ip
}

func (f *FileName) GetDate() string {
	return f.Name[0:8]
}

type FileNameParser interface {
	GetLocalIP()
	GetDate()
}
