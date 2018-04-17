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
	"strconv"
	"testing"
	"time"

	"github.com/m-lab/etl-embargo"
)

func TestLoadWhitelist(t *testing.T) {
	embargo_check := new(embargo.EmbargoCheck)
	if !embargo_check.LoadWhitelist() {
		t.Error("Do not load site IP json file from cloud storage correctly.")
	}
	return
}

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

func TestCheckInWhitelist(t *testing.T) {
	embargo_check := new(embargo.EmbargoCheck)
	embargo_check.ReadWhitelistFromLocal("testdata/whitelist")
	// After embargo date and IP not whitelisted. Return true, embargoed.
	filename1 := "20180225T23:00:00Z_4.34.58.34_0.web100.gz"
	if embargo_check.CheckInWhitelist(filename1) {
		t.Errorf("CheckInWhitelist(%s) = true, but IP not whitelisted (%v).\n", filename1, embargo_check.Whitelist)
	}

	// IP whitelisted. Return false, not embargoed.
	filename2 := "20170225T23:00:00Z_213.244.128.170_0.web100.gz"
	if !embargo_check.CheckInWhitelist(filename2) {
		t.Errorf("CheckInWhitelist(%s) = false, but IP whitelisted (%v).\n", filename2, embargo_check.Whitelist)
	}
	return
}

func TestCheckWhetherMoreThanOneYearOld(t *testing.T) {
	currentTime, _ := strconv.Atoi(time.Now().UTC().Format("20061229"))
	if embargo.CheckWhetherMoreThanOneYearOld(currentTime, 0) {
		t.Error("The current date should return false for WhetherMoreThanOneYearOld check.")
	}
	if !embargo.CheckWhetherMoreThanOneYearOld(20060129, 0) {
		t.Error("This last year date should return true for WhetherMoreThanOneYearOld check.")
	}
	if embargo.CheckWhetherMoreThanOneYearOld(20170329, 20160829) {
		t.Error("This input date 20170329 should return false for WhetherMoreThanOneYearOld check given cutoff date 20160829.")
	}
}

func TestGetDayOfWeek(t *testing.T) {
	dayOfWeek, err := embargo.GetDayOfWeek("sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz")
	if err != nil || dayOfWeek != "Tuesday" {
		t.Error("Did not get day of week correctly.\n")
	}
}
