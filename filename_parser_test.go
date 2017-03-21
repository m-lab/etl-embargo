package embargo

import (
	"fmt"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
  result1 := GetLocalIP("20170225T23:00:00Z_4.34.58.34_0.web100.gz")
  if result1 != "4.34.58.34" {
    t.Errorf("wrong! %v\n", result1)
    return
  }

  result2 := GetLocalIP("20170225T23:00:00Z_ALL0.web100.gz")
  if result2 != "" {
    t.Errorf("wrong! %v\n", result2)
    return
  }
  fmt.Printf("GetLocalIP Correct!\n")
}

func TestGetDate(t *testing.T) {
  result1 := GetDate("20170225T23:00:00Z_4.34.58.34_0.web100.gz")
  if result1 != "20170225" {
    t.Errorf("wrong! %v\n", result1)
    return
  }
}
