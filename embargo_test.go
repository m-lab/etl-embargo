package embargo_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	embargo "github.com/m-lab/etl-embargo"
)

func TestEmbargo(t *testing.T) {
	sourceBucket := "embargo-source-mlab-testing"
	publicBucket := "embargo-output-mlab-testing"
	privateBucket := "embargoed-data-mlab-testing"
	testConfig := embargo.NewEmbargoConfig(sourceBucket, privateBucket, publicBucket, "testdata/whitelist_full")
	embargo.DeleteFiles(sourceBucket, "")
	embargo.UploadFile(sourceBucket, "testdata/20170315T000000Z-mlab3-sea03-sidestream-0000.tgz", "sidestream/2017/03/15/")
	if testConfig.EmbargoOneDayData("20170315", 20160822) != nil {
		t.Error("Did not perform embargo correctly.\n")
	}

	// Verify that there are expected outputs in the destination buckets.
	if !embargo.CompareBuckets(privateBucket, "embargoed-golden-data-mlab-testing") {
		t.Error("Did not generate embargoed data correctly.\n")
	}
	if !embargo.CompareBuckets(publicBucket, "embargo-output-golden-mlab-testing") {
		t.Error("Did not generate public data correctly.\n")
	}

	// Cleanup
	embargo.DeleteFiles(sourceBucket, "")
	embargo.DeleteFiles(privateBucket, "")
	embargo.DeleteFiles(publicBucket, "")
	return
}

// This test verifies that func SplitFile() correctly splits the input tar
// file into 2 tar files: one contains the embargoed web100 files, the other
// contains the files that can be published.
// TODO: a cleaner way to test this would be to create a tar file on the fly,
// with lists of inner files, call SplitFile on it, then verify that the pub
// and private buffers contain the correct filenames.
func TestSplitTarFile(t *testing.T) {
	testConfig := embargo.NewEmbargoConfig("embargo-source-mlab-testing", "embargoed-data-mlab-testing", "embargo-output-mlab-testing", "testdata/whitelist_full")

	// Load input tar file.
	file, err := os.Open("testdata/20170315T000000Z-mlab3-sea03-sidestream-0000.tgz")
	if err != nil {
		t.Fatal("cannot open test data.")
	}
	defer file.Close()

	privateBuf, publicBuf, err := testConfig.SplitFile(file, false)
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
