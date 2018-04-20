package embargo_test

import (
	"fmt"
	"testing"

	embargo "github.com/m-lab/etl-embargo"
)

func TestBucketOperations(t *testing.T) {
	destBucket := "bucket-gcs-operations-mlab-testing"
	sourceBucket := "embargo-mlab-testing"

	result := embargo.CopyOneFile(sourceBucket, destBucket, "whitelist_full")
	if result == false {
		t.Errorf("Cannot copy file from another bucket.")
		return
	}

	fileNames := embargo.GetFileNamesFromBucket(destBucket)

	fmt.Printf("Files in bucket %v:\n", destBucket)
	for _, fileName := range fileNames {
		fmt.Println(fileName)
	}

	result = embargo.CompareBuckets(destBucket, sourceBucket)
	if result == false {
		t.Errorf("The two buckets are not the same.")
		return
	}
	result = embargo.DeleteFiles(destBucket, "")
	if result == false {
		t.Errorf("Cannot delete files. The bucket bucket-gcs-operations-mlab-testing needs to be cleaned up before rerunning the test.")
		return
	}
}
