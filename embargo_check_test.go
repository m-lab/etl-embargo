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

package main

import (
	"testing"
)

func TestReadWhitelistFromLocal(t *testing.T) {
	embargo_check := new(EmbargoCheck)
	embargo_check.ReadWhitelistFromLocal("testdata/whitelist")
	if !embargo_check.Whitelist["213.244.128.170"] {
		t.Error("missing IP in Whitelist: want '213.244.128.170'\n")
	}
	if embargo_check.Whitelist["2001:4c08:2003:2::16"] {
		t.Error("IP 2001:4c08:2003:2::16 should not be in Whitelist.\n")
	}
	return
}

/*
// Require authentication to run.
func TestReadWhitelistFromGCS(t *testing.T) {
	embargo_check := new(EmbargoCheck)
	embargo_check.ReadWhitelistFromGCS("whitelist")
	if !embargo_check.Whitelist["213.244.128.170"] {
		t.Error("missing IP in Whitelist: want '213.244.128.170'\n")
	}
	if embargo_check.Whitelist["2001:4c08:2003:2::16"] {
		t.Error("IP 2001:4c08:2003:2::16 should not be in Whitelist.\n")
	}
	return
}
*/
func TestShouldEmbargo(t *testing.T) {
	embargo_check := new(EmbargoCheck)
	embargo_check.ReadWhitelistFromLocal("testdata/whitelist")
	embargo_check.Embargodate = 20160315
	// After embargo date and IP not whitelisted. Return true, embargoed.
	filename1 := "20170225T23:00:00Z_4.34.58.34_0.web100.gz"
	if !embargo_check.ShouldEmbargo(filename1) {
		t.Error("ShouldEmbargo(%s) = false, but file date is after embargo date (%d) and IP not whitelisted (%v).\n", filename1, embargo_check.Embargodate, embargo_check.Whitelist)
	}

	// After embargo date and IP whitelisted. Return false, not embargoed.
	filename2 := "20170225T23:00:00Z_213.244.128.170_0.web100.gz"
	if embargo_check.ShouldEmbargo(filename2) {
		t.Error("ShouldEmbargo(%s) = true, but after embargo date (%d) and IP whitelisted (%v).\n", filename2, embargo_check.Embargodate, embargo_check.Whitelist)
	}
	// Before embargo date. Return false, not embargoed
	filename3 := "20150225T23:00:00Z_213.244.128.1_0.web100.gz"
	if embargo_check.ShouldEmbargo(filename3) {
		t.Error("ShouldEmbargo(%s) = true, but before embargo date(%d).\n", filename3, embargo_check.Embargodate)
	}
	return
}
