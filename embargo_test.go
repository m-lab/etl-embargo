package embargo_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	embargo "github.com/m-lab/etl-embargo"
)

// End to end test, requires authentication.
// TODO: Enable it on Travis.
func TestEmbargo(t *testing.T) {
	sourceBucket := "embargo-test"
	testConfig := embargo.NewEmbargoConfig(sourceBucket, "mlab-embargoed-data", "embargo-output", "")
	embargo.DeleteFiles(sourceBucket, "")
	embargo.UploadFile(sourceBucket, "testdata/20170315T000000Z-mlab3-sea03-sidestream-0000.tgz", "sidestream/2017/03/15/")
	if testConfig.EmbargoOneDayData("2017/03/15") != nil {
		t.Error("Did not perform embargo correctly.\n")
	}
	embargo.DeleteFiles(sourceBucket, "")
	return
}

func TestGetDayOfWeek(t *testing.T) {
	dayOfWeek, err := embargo.GetDayOfWeek("sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz")
	if err != nil || dayOfWeek != "Tuesday" {
		t.Error("Did not get day of week correctly.\n")
	}
}

// This test verifies that func embargoBuf() correctly splits the input tar
// file into 2 tar files: one contains the embargoed web100 files, the other
// contains the files that can be published.
// TODO: a cleaner way to test this would be to create a tar file on the fly,
// with lists of inner files, call SplitFile on it, then verify that the pub
// and private buffers contain the correct filenames.
func TestSplitTarFile(t *testing.T) {
	testConfig := embargo.NewEmbargoConfig("sidestream-embargo", "mlab-embargoed-data", "embargo-output", "")

	// Load input tar file.
	file, err := os.Open("testdata/20170315T000000Z-mlab3-sea03-sidestream-0000.tgz")
	if err != nil {
		t.Fatal("cannot open test data.")
	}
	defer file.Close()

	privateBuf, publicBuf, err := testConfig.SplitFile(file, 20160820)
	if err != nil {
		t.Error("Did not perform embargo correctly.\n")
	}
	publicGolden, err := os.Open("testdata/20170315T000000Z-mlab3-sea03-sidestream-0000-p.tgz")
	if err != nil {
		t.Fatal("cannot open public golden data.")
	}
	defer publicGolden.Close()
	publicContent, _ := ioutil.ReadAll(publicGolden)
	if !bytes.Equal(publicBuf.Bytes(), publicContent) {
		t.Error("Public data not correct.\n")
	}

	privateGolden, err := os.Open("testdata/20170315T000000Z-mlab3-sea03-sidestream-0000-e.tgz")
	if err != nil {
		t.Fatal("cannot open private golden data.")
	}
	defer privateGolden.Close()
	privateContent, _ := ioutil.ReadAll(privateGolden)
	if !bytes.Equal(privateBuf.Bytes(), privateContent) {
		t.Error("Private data not correct.\n")
	}
}
