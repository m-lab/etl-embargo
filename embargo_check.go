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
	"errors"
	"log"
	"os"
	"strings"
	"time"
)

type EmbargoCheck struct {
	Whitelist map[string]bool
}

// Given the current date, return true if the date is earlier than the cutoffDate.
// The input date is integer in format yyyymmdd
// If the input cutoffDate is 0, use one year ago of currentTime.
func CheckWhetherMoreThanOneYearOld(date int, cutoffDate int) bool {
	currentTime := time.Now()
	if cutoffDate == 0 {
		cutoffDate = (currentTime.Year()-1)*10000 + int(currentTime.Month())*100 + currentTime.Day()
	}
	if date < cutoffDate {
		return true
	}
	return false
}

// For a filepath string like
// "sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz",
// return "Tuesday" for date "2017/05/16"
func GetDayOfWeek(filename string) (string, error) {
	if len(filename) < 21 {
		return "", errors.New("invalid filename.")
	}
	date := filename[11:21]
	dateStr := strings.Replace(date, "/", "-", -1) + " 00:00:00"
	parsedDate, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return "", err
	}
	return parsedDate.Weekday().String(), nil
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

// Check whether a file with IP in the whitelist.
// Always return true for non-web100 files.
// The filename is like: 20170225T23:00:00Z_4.34.58.34_0.web100
// file with IP that is in the IP whitelist, return true
// file with IP not in the whitelist, return false
func (ec *EmbargoCheck) CheckInWhitelist(fileName string) bool {
	if !strings.Contains(fileName, "web100") {
		return true
	}

	fn := FileName{Name: fileName}
	localIP := fn.GetLocalIP()
	if ec.Whitelist[localIP] {
		return true
	}
	return false
}
