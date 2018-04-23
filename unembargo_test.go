package embargo_test

import (
	"testing"

	"github.com/m-lab/etl-embargo"
)

// This end to end test require anthentication for running on Travis.
func TestUnembargoLegacy(t *testing.T) {
	privateBucket := "unembargoed-data-mlab-testing"
	publicBucket := "bigstore-data-mlab-testing"
	testConfig := embargo.NewConfig(privateBucket, publicBucket)
	// Prepare the buckets for input & output.
	embargo.DeleteFiles(privateBucket, "")
	embargo.UploadFile(privateBucket, "testdata/20160102T000000Z-mlab3-sin01-sidestream-0000.tgz", "sidestream/2016/01/02/")
	embargo.DeleteFiles(publicBucket, "")
	if testConfig.Unembargo(20160102) != nil {
		t.Errorf("Unembargo func did not return true.")
		return
	}
	// Check the privateBucket does not have that file any more
	privateNames := embargo.GetFileNamesFromBucket(privateBucket)
	if len(privateNames) != 0 {
		t.Errorf("The embargoed copy should be deleted after the process.\n")
	}
	// Check the publicBucket has that file
	publicNames := embargo.GetFileNamesFromBucket(publicBucket)
	if len(publicNames) != 1 || publicNames[0] != "sidestream/2016/01/02/20160102T000000Z-mlab3-sin01-sidestream-0000.tgz" {
		t.Errorf("The public bucket does not have the new copy.\n")
	}

}
