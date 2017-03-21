package embargo

import (
	"fmt"
	"testing"
)

func TestReadWhitelistFromLocal(t *testing.T) {
  white_list := ReadWhitelistFromLocal("whitelist")
  if white_list["213.244.128.170"] {
    fmt.Printf("ReadWhitelist correct\n")
  } else {
    t.Error("Wrong\n")
  }
  if white_list["2001:4c08:2003:2::16"] {
    t.Error("Wrong\n")
  } else {
    fmt.Printf("ReadWhitelist correct\n")
  }
  return
}

func TestReadWhitelistFromGCS(t *testing.T) {
  white_list := ReadWhitelistFromGCS("whitelist")
  if white_list["213.244.128.170"] {
    fmt.Printf("ReadWhitelist correct\n")
  } else {
    t.Error("Wrong\n")
  }
  if white_list["2001:4c08:2003:2::16"] {
    t.Error("Wrong\n")
  } else {
    fmt.Printf("ReadWhitelist correct\n")
  }
  return
}

func TestShouldEmbargo(t *testing.T) {
  whitelist := ReadWhitelistFromLocal("whitelist")
  embargoDate    = "20160315"
  // After embargo date and IP not whitelisted. Return true, embargoed 
  if ShouldEmbargo("20170225T23:00:00Z_4.34.58.34_0.web100.gz", whitelist) {
    fmt.Printf("Embargo correctly.\n")
  } else {
    t.Error("Wrong. After embargo date and IP not whitelisted, should be embargoed.\n")
  }

  // After embargo date and IP whitelisted. Return false, not embargoed 
  if !ShouldEmbargo("20170225T23:00:00Z_213.244.128.170_0.web100.gz", whitelist) {
    fmt.Printf("Embargo correctly.\n")
  } else {
    t.Error("Wrong. After embargo data and IP whitelisted, should not be embargoed.\n")
  }
  // Before embargo date. Return false, not embargoed 
  if !ShouldEmbargo("20150225T23:00:00Z_213.244.128.1_0.web100.gz", whitelist) {
    fmt.Printf("Embargo correctly.\n")
  } else {
    t.Error("Wrong. Before embargo date, should not be embargoed.\n")
  }
  return
}

func TestEmbargo(t *testing.T) {
  embargoDate    = "20160315"
  sourceBucket   = "sidestream-embargo"
  destBucket     = "embargo-output"
  if Embargo() {
    fmt.Printf("Embargo correct\n")
  } else {
    t.Error("wrong\n")
  }
  return
}

