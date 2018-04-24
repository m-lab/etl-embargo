package embargo_test

import (
	"testing"

	"github.com/m-lab/etl-embargo"
)

func TestReadSiteIPlistFromLocal(t *testing.T) {
	embargo_check := new(embargo.EmbargoCheck)
	embargo_check.ReadSiteIPlistFromLocal("testdata/whitelist")
	if !embargo_check.SiteIPList["213.244.128.170"] {
		t.Error("missing IP in site IP list: want '213.244.128.170'\n")
	}
	if embargo_check.SiteIPList["2001:4c08:2003:2::16"] {
		t.Error("IP 2001:4c08:2003:2::16 should not be in site IP list.\n")
	}
	return
}

func TestParseJson(t *testing.T) {
	body := []byte(`[
  {
    "hostname": "mlab2.samknows.acc02.measurement-lab.org", 
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
	siteIPList, err := embargo.ParseJson(body)
	if err != nil || len(siteIPList) != 2 || !siteIPList["196.49.14.227"] || siteIPList["196.49.14.214"] {
		t.Error("Do not parse bytes into struct correctly.")
	}
}

func TestCheckInSiteIPList(t *testing.T) {
	embargo_check := new(embargo.EmbargoCheck)
	embargo_check.ReadSiteIPlistFromLocal("testdata/whitelist")
	// After embargo date and IP not in site IP list. Return true, embargoed.
	filename1 := "20180225T23:00:00Z_4.34.58.34_0.web100.gz"
	if embargo_check.CheckInSiteIPList(filename1) {
		t.Errorf("CheckInSiteIPList(%s) = true, but IP not in site IP list (%v).\n", filename1, embargo_check.SiteIPList)
	}

	// IP in site IP list. Return false, not embargoed.
	filename2 := "20170225T23:00:00Z_213.244.128.170_0.web100.gz"
	if !embargo_check.CheckInSiteIPList(filename2) {
		t.Errorf("CheckInSiteIPList(%s) = false, but IP in site IP list (%v).\n", filename2, embargo_check.SiteIPList)
	}
	return
}

func TestGetDayOfWeek(t *testing.T) {
	dayOfWeek, err := embargo.GetDayOfWeek("sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz")
	if err != nil || dayOfWeek != "Tuesday" {
		t.Error("Did not get day of week correctly.\n")
	}
}
