package embargo_test

import (
	"fmt"
	"testing"

	embargo "github.com/m-lab/etl-embargo"
)

func TestBucketOperations(t *testing.T) {
	bucketName := "bucket-gcs-operations-mlab-testing"
	sourceBucket := "embargo-mlab-testing"

	result := embargo.CopyOneFile(sourceBucket, bucketName, "whitelist_full")
	if result == false {
		t.Errorf("Cannot copy file from another bucket.")
		return
	}

	fileNames := embargo.GetFileNamesFromBucket(bucketName)

	fmt.Printf("Files in bucket %v:\n", bucketName)
	for _, fileName := range fileNames {
		fmt.Println(fileName)
	}

	result = embargo.CompareBuckets(bucketName, sourceBucket)
	if result == false {
		t.Errorf("The two buckets are not the same.")
		return
	}

	result = embargo.DeleteFiles(bucketName, "")
	if result == false {
		t.Errorf("Cannot delete files.")
		return
	}
}
