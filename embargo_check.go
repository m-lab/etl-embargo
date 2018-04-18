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

// Implement whitelist loading and embargo check based on filename.
package embargo

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type EmbargoCheck struct {
	Whitelist map[string]bool
}

// Load whitelist IP info from cloud storage.
const SITE_IP_URL_TEST = "https://storage.googleapis.com/operator-mlab-staging/metadata/v0/current/mlab-host-ips.json"
const SITE_IP_URL = "https://storage.googleapis.com/operator-mlab-oti/metadata/v0/current/mlab-host-ips.json"

type Site struct {
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Ipv6     string `json:"ipv6"`
}

// ParseJson parses bytes into array of struct.
func ParseJson(body []byte) (map[string]bool, error) {
	sites := make([]Site, 0)
	whiteList := make(map[string]bool)
	if err := json.Unmarshal(body, &sites); err != nil {
		log.Printf("Cannot parse site IP json files.")
		return whiteList, errors.New("Cannot parse site IP json files.")
	}

	for _, site := range sites {
		if site.Ipv4 != "" {
			whiteList[site.Ipv4] = true
		}
		if site.Ipv6 != "" {
			whiteList[site.Ipv6] = true
		}
	}
	return whiteList, nil
}

// LoadWhitelist load the IP whitelist from GCS.
func (ec *EmbargoCheck) LoadWhitelist() bool {
	project := os.Getenv("GCLOUD_PROJECT")
	log.Printf("Using project: %s\n", project)
	json_url := SITE_IP_URL_TEST
	if project == "mlab-oti" {
		json_url = SITE_IP_URL
	}

	resp, err := http.Get(json_url)
	if err != nil {
		log.Printf("cannot download site IP json file.\n")
		return false
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Cannot read site IP json files.\n")
		return false
	}

	ec.Whitelist, err = ParseJson(body)
	if err == nil {
		return true
	}
	return false
}

// ReadWhitelistFromLocal load IP whitelist from a local file.
func (ec *EmbargoCheck) ReadWhitelistFromLocal(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	whiteList := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		oneLine := strings.TrimSuffix(scanner.Text(), "\n")
		whiteList[oneLine] = true
	}
	ec.Whitelist = whiteList
	return true
}

// ReadWhitelistFromGCS load IP whitelist from cloud storage.
func (ec *EmbargoCheck) ReadWhitelistFromGCS(bucket string, path string) bool {
	// TODO: Create service in a Singleton object, and reuse them for all GCS requests.
	checkService := CreateService()
	if checkService == nil {
		log.Printf("Storage service was not initialized.\n")
		return false
	}
	whiteList := make(map[string]bool)
	if fileContent, err := checkService.Objects.Get(bucket, path).Download(); err == nil {
		scanner := bufio.NewScanner(fileContent.Body)
		for scanner.Scan() {
			oneLine := strings.TrimSuffix(scanner.Text(), "\n")
			whiteList[oneLine] = true
		}
		ec.Whitelist = whiteList
		return true
	}
	return false
}

// EmbargoCheck decide whether to embargo it based on embargo date and IP
// whitelist given a filename of sidestream test.
// Always return false for non-web100 files.
// The filename is like: 20170225T23:00:00Z_4.34.58.34_0.web100
// THe embargo date is like 20160225
// file with date on or before the embargo date are always published. Return false
// file with IP that is in the IP whitelist are always published. Return false
// file with date after the embargo date and IP not in the whitelist will be embargoed. Return true
// For old file format like 2017/03/15/mlab3.sea03/20170315T12:00:00Z_ALL0.web100
// it will return true always.
func (ec *EmbargoCheck) ShouldEmbargo(fileName string) bool {
	if !strings.Contains(fileName, "web100") {
		return false
	}
	if len(fileName) < 8 {
		log.Println("Filename not with right length.")
		return true
	}
	date, err := strconv.Atoi(fileName[0:8])
	if err != nil {
		log.Println(err)
		return true
	}

	// CheckWhetherUnembargo(date) return true if this date is more than one year old.
	if CheckWhetherUnembargo(date) {
		return false
	}
	fn := FileName{Name: fileName}
	localIP := fn.GetLocalIP()
	if ec.Whitelist[localIP] {
		return false
	}
	return true
}
