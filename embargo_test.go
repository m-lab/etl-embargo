package embargo_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	embargo "github.com/m-lab/etl-embargo"
)

func cleanUpBucket(bucketName string) {
	if !embargo.DeleteFiles(bucketName, "") {
		fmt.Printf("Delete file failed, Please delete files in %s before rerunning the test.\n", bucketName)
	}
}

func TestEmbargo(t *testing.T) {
	testConfig, err := embargo.GetEmbargoConfig("testdata/whitelist_full")
	if err != nil {
		t.Error(err.Error())
		return
	}
	sourceBucket := "scraper-mlab-testing"
	privateBucket := "embargo-mlab-testing"
	publicBucket := "archive-mlab-testing"
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

	cleanUpBucket(sourceBucket)
	cleanUpBucket(privateBucket)
	cleanUpBucket(publicBucket)
	return
}

// This test verifies that func SplitFile() correctly splits the input tar
// file into 2 tar files: one contains the embargoed web100 files, the other
// contains the files that can be published.
// TODO: a cleaner way to test this would be to create a tar file on the fly,
// with lists of inner files, call SplitFile on it, then verify that the pub
// and private buffers contain the correct filenames.
func TestSplitTarFile(t *testing.T) {
	testConfig, err := embargo.GetEmbargoConfig("testdata/whitelist_full")
	if err != nil {
		t.Error(err.Error())
		return
	}
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
