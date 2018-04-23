package embargo_test

import (
	"testing"

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

func TestGetDayOfWeek(t *testing.T) {
	dayOfWeek, err := embargo.GetDayOfWeek("sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz")
	if err != nil || dayOfWeek != "Tuesday" {
		t.Error("Did not get day of week correctly.\n")
	}
}
