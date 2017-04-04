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

package embargo

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// This end to end test require anthentication.
func TestUnembargoLegacy(t *testing.T) {
	privateBucket = "mlab-embargoed-data"
	publicBucket = "mlab-bigstore-data"
	// Prepare the buckets for input & output.
	DeleteFiles(privateBucket, "")
	UploadFile(privateBucket, "testdata/20160102T000000Z-mlab3-sin01-sidestream-0000.tgz", "sidestream/2016/01/02/")
	DeleteFiles(publicBucket, "")
	if Unembargo(20160102) {
		// Check the privateBucket does not have that file any more
		fileNames := GetFileNamesFromBucket(privateBucket)
		for _, fileName := range fileNames {
			if strings.Contains(fileName, "20160102T000000Z-mlab3-sin01-sidestream-0000.tgz") {
				t.Errorf("The embargoed copy should be deleted after the process.\n")
			}
		}
		// Check the publicBucket has that file
		fileNames2 := GetFileNamesFromBucket(publicBucket)
		for _, fileName2 := range fileNames2 {
			if strings.Contains(fileName2, "20160102T000000Z-mlab3-sin01-sidestream-0000.tgz") {
				return
			}
		}
		t.Errorf("The public bucket does not have the new copy.\n")
	} else {
		t.Errorf("Unembargo func did not return true.")
	}
}

func TestCalculateDate(t *testing.T) {
	current_time, _ := strconv.Atoi(time.Now().UTC().Format("20061229"))
	if CheckWhetherUnembargo(current_time) {
		t.Error("The current date should return false for unembargo check.")
	}
	if CheckWhetherUnembargo(20161102) {
		t.Error("The current date should return false for unembargo check.")
	}
	if !CheckWhetherUnembargo(20060129) {
		t.Error("This last year date should return true for unembargo check.")
	}
}
