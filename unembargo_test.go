package embargo

import (
	"fmt"
	"testing"
)

/*
// This end to end test require anthentication
func TestUnembargoLegacy(t *testing.T) {
	privateBucket = "mlab-embargoed-data"
	publicBucket = "mlab-bigstore-data"
	UnEmbargoOneDayLegacyFiles(privateBucket, publicBucket, "sidestream/2016/11/02")
}
*/

func TestCalculateDate(t *testing.T) {
	fmt.Println("This is the date one year ago.")
	fmt.Println(CalculateUnembargoDate())
}
