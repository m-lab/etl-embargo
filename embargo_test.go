package embargo

import (
	"bytes"
        "fmt"
	"io/ioutil"
	"os"
	"testing"
)

/*
// End to end test, requires authentication.
func TestEmbargo(t *testing.T) {
	embargoDate = 20160315
	sourceBucket = "sidestream-embargo"
	destPublicBucket = "embargo-output"
	destPrivateBucket = "mlab-embargoed-data"
	if !Embargo() {
		t.Error("Did not perform embargo ocrrectly.\n")
	}
	return
}
*/

func TestSplitTarFile(t *testing.T) {
	embargoCheck.ReadWhitelistFromLocal("testdata/whitelist_full")
	embargoCheck.Embargodate = 20160315
	// Load input tar file.
	file, err := os.Open("testdata/20170315T000000Z-mlab3-sea03-sidestream-0000.tgz")
	if err != nil {
		t.Fatal("cannot opene test data.")
	}
	defer file.Close()

	suc, privateBuf, publicBuf := SplitFile(file)
	if !suc {
		t.Error("Did not perform embargo ocrrectly.\n")
	}
	publicGolden, err := os.Open("testdata/20170315T000000Z-mlab3-sea03-sidestream-0000-p.tgz")
	if err != nil {
		t.Fatal("cannot opene public golden data.")
	}
	defer publicGolden.Close()
	publicContent, err := ioutil.ReadAll(publicGolden)
	if !bytes.Equal(publicBuf.Bytes(), publicContent) {
		t.Error("Public data not correct.\n")
	}

	privateGolden, err := os.Open("testdata/20170315T000000Z-mlab3-sea03-sidestream-0000-e.tgz")
	if err != nil {
		t.Fatal("cannot opene private golden data.")
	} else {
                fmt.Println("correct")
        }
	defer privateGolden.Close()
	privateContent, err := ioutil.ReadAll(privateGolden)
	if !bytes.Equal(privateBuf.Bytes(), privateContent) {
		t.Error("Private data not correct.\n")
	}
}
