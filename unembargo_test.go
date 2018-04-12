/*
Copyright 2013 Google Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package embargo_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/m-lab/etl-embargo"
)

// This end to end test require anthentication.
func TestUnembargoLegacy(t *testing.T) {
	privateBucket := "mlab-embargoed-data-test"
	publicBucket := "mlab-bigstore-data-test"
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

func TestCalculateDate(t *testing.T) {
	currentTime, _ := strconv.Atoi(time.Now().UTC().Format("20061229"))
	if embargo.CheckWhetherMoreThanOneYearOld(currentTime, 0) {
		t.Error("The current date should return false for unembargo check.")
	}
	if !embargo.CheckWhetherMoreThanOneYearOld(20060129, 0) {
		t.Error("This last year date should return true for unembargo check.")
	}
}
