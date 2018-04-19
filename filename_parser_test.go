package embargo_test

import (
	"testing"

	embargo "github.com/m-lab/etl-embargo"
)

func TestGetLocalIP(t *testing.T) {
	fn1 := embargo.FileName{Name: "20170225T23:00:00Z_4.34.58.34_0.web100.gz"}
	if fn1.GetLocalIP() != "4.34.58.34" {
		t.Errorf("Wrong!\n")
		return
	}

	fn2 := embargo.FileName{Name: "20170225T23:00:00Z_ALL0.web100.gz"}
	if fn2.GetLocalIP() != "" {
		t.Errorf("Wrong!\n")
		return
	}

	fn3 := embargo.FileName{Name: "20170225T23:00:00Z_2001:4c08:2003:3f:::230_ALL0.web100.gz"}
	if fn3.GetLocalIP() != "2001:4c08:2003:3f::230" {
		t.Errorf("Wrong!\n")
		return
	}
}

func TestGetDate(t *testing.T) {
	fn1 := embargo.FileName{Name: "20170225T23:00:00Z_4.34.58.34_0.web100.gz"}
	if fn1.GetDate() != "20170225" {
		t.Errorf("Wrong!\n")
		return
	}
}
