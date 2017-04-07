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
	Whitelist   map[string]bool
	Embargodate int
}

// TODO: Read IP whitelist from Data Store.

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
func (ec *EmbargoCheck) ReadWhitelistFromGCS(path string) bool {
	// TODO: Create service in a Singleton object, and reuse them for all GCS requests.
	checkService := CreateService()
	if checkService == nil {
		log.Printf("Storage service was not initialized.\n")
		return false
	}
	whiteList := make(map[string]bool)
	if fileContent, err := checkService.Objects.Get("sidestream-embargo", path).Download(); err == nil {
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
// The filename is like: 20170225T23:00:00Z_4.34.58.34_0.web100
// THe embargo date is like 20160225
// file with date on or before the embargo date are always published. Return false
// file with IP that is in the IP whitelist are always published. Return false
// file with date after the embargo date and IP not in the whitelist will be embargoed. Return true
// For old file format like 2017/03/15/mlab3.sea03/20170315T12:00:00Z_ALL0.web100
// it will return true always.
func (ec *EmbargoCheck) ShouldEmbargo(fileName string) bool {
	if len(fileName) < 8 {
		log.Println("Filename not with right length.\n")
		return true
	}
	date, err := strconv.Atoi(fileName[0:8])
	if err != nil {
		log.Println(err)
		return true
	}

	if err != nil {
		log.Println(err)
		return true
	}
	if date < ec.Embargodate {
		return false
	}
	fn := FileName{name: fileName}
	localIP := fn.GetLocalIP()
	if ec.Whitelist[localIP] {
		return false
	}
	return true
}
