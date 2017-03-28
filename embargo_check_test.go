package embargo

import (
	"fmt"
	"testing"
)
/*
func TestReadWhitelistFromLocal(t *testing.T) {
  embargo_check := new(EmbargoCheck)
  embargo_check.ReadWhitelistFromLocal("whitelist")
  if embargo_check.whitelist["213.244.128.170"] {
    fmt.Printf("ReadWhitelist correct\n")
  } else {
    t.Error("Wrong\n")
  }
  if embargo_check.whitelist["2001:4c08:2003:2::16"] {
    t.Error("Wrong\n")
  } else {
    fmt.Printf("ReadWhitelist correct\n")
  }
  return
}

// Requite authentication to run this test.
func TestReadWhitelistFromGCS(t *testing.T) {
  embargo_check := new(EmbargoCheck)
  embargo_check.ReadWhitelistFromGCS("whitelist")
  if embargo_check.whitelist["213.244.128.170"] {
    fmt.Printf("ReadWhitelist correct\n")
  } else {
    t.Error("Wrong\n")
  }
  if embargo_check.whitelist["2001:4c08:2003:2::16"] {
    t.Error("Wrong\n")
  } else {
    fmt.Printf("ReadWhitelist correct\n")
  }
  return
}

func TestShouldEmbargo(t *testing.T) {
  embargo_check := new(EmbargoCheck)
  embargo_check.ReadWhitelistFromLocal("whitelist")
  embargo_check.embargodate = "20160315"
  // After embargo date and IP not whitelisted. Return true, embargoed 
  if embargo_check.ShouldEmbargo("20170225T23:00:00Z_4.34.58.34_0.web100.gz") {
    fmt.Printf("Embargo correctly.\n")
  } else {
    t.Error("Wrong. After embargo date and IP not whitelisted, should be embargoed.\n")
  }

  // After embargo date and IP whitelisted. Return false, not embargoed 
  if !embargo_check.ShouldEmbargo("20170225T23:00:00Z_213.244.128.170_0.web100.gz") {
    fmt.Printf("Embargo correctly.\n")
  } else {
    t.Error("Wrong. After embargo data and IP whitelisted, should not be embargoed.\n")
  }
  // Before embargo date. Return false, not embargoed 
  if !embargo_check.ShouldEmbargo("20150225T23:00:00Z_213.244.128.1_0.web100.gz") {
    fmt.Printf("Embargo correctly.\n")
  } else {
    t.Error("Wrong. Before embargo date, should not be embargoed.\n")
  }
  return
}
*/
