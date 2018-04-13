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

func TestReadWhitelistFromGCS(t *testing.T) {
	embargo_check := new(embargo.EmbargoCheck)
	embargo_check.ReadWhitelistFromGCS("embargo-test", "whitelist_full")
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

func TestCalculateDate(t *testing.T) {
	currentTime, _ := strconv.Atoi(time.Now().UTC().Format("20061229"))
	if embargo.CheckWhetherMoreThanOneYearOld(currentTime, 0) {
		t.Error("The current date should return false for unembargo check.")
	}
	if !embargo.CheckWhetherMoreThanOneYearOld(20060129, 0) {
		t.Error("This last year date should return true for unembargo check.")
	}
}
