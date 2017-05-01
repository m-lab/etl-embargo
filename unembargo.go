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

// Implement the umembargo process when the previously embargoed files are more than one year old.
package embargo

// For example, the legacy pipeline stop running from 2017-05-01
// There are three kinds of files which require umembargo operations:
// 1. the legacy sidestream files using the old filename format which do not have IP address.
//    Those files are less than one years old when the old pipeline would retire.
//    For the above example, the files from 2016-05-01 till 2017-03-01
// Solutions: We have convert those archived tests from inside-Google format to tar format which matches
//            the files generated by new scraper. So the only operation is to replace the current
//            files in the BigStore using the converted tar files.
//
//  2. the sidestream files using the new format, but not one year old yet.
//     The files from 2017-03-01 to 2017-05-01
//  Solutions: Those files have been embargoed in legacy pipeline. Using the file content instead of
//             IP address filename. We could just replace the current files in the BigStore using
//             the converted tar files.
//
//  3. the new incoming sidestream files after 2017-05-01
//     Those files have been embargoed by the new process using the filename which split the files
//     into two tars: one goes to public BigStore bucket, and one hold in some private bucket.
//  Solutions: We need to move the previously embargoed files to public bucket, make sure they
//             are added with a different name, and not coving the previous public ones.
//
// Since during embargo process of new platform ,the embargoed private files are named
// differently with public files, we can use the same function to handle all above 3 conditions:
// 1. Cover the old public file with the new one if the name is the same.
// 2. Copy the private files directlyif there is no existing public files with the same name.

import (
	"fmt"
	"golang.org/x/net/context"
	storage "google.golang.org/api/storage/v1"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	privateBucket = ""
	publicBucket  = ""
)

// Given the current date, return true if the date is more than oneyear ago.
// The input date is integer in format yyyymmdd
func CheckWhetherUnembargo(date int) bool {
	current_time := time.Now()
	cutoff_date := (int(current_time.Year()) -1) * 10000 + int(current_time.Month()) * 100 + int(current_time.Day())
	if date < cutoff_date {
		return true
	}
	return false
}

// Get filenames for given bucket with the given prefix. Use the service
func GetFileNamesWithPrefix(service *storage.Service, bucketName string, prefixFileName string) (map[string]bool, error) {
	existingFilenames := make(map[string]bool)
	destPageToken := ""
	for {
		destinationFiles := service.Objects.List(bucketName)
		if destPageToken != "" {
			destinationFiles.PageToken(destPageToken)
		}
		destinationFiles.Prefix(prefixFileName)
		destinationFilesList, err := destinationFiles.Context(context.Background()).Do()
		if err != nil {
			log.Printf("Objects.List failed: %v\n", err)
			return existingFilenames, err
		}
		for _, oneItem := range destinationFilesList.Items {
			existingFilenames[oneItem.Name] = true
		}
		destPageToken = destinationFilesList.NextPageToken
		if destPageToken == "" {
			break
		}
	}
	return existingFilenames, nil
}

// The date is used as prefixFileName in format sidestream/yyyy/mm/dd
func UnEmbargoOneDayLegacyFiles(sourceBucket string, destBucket string, prefixFileName string) error {
	unembargoService := CreateService()
	if unembargoService == nil {
		log.Printf("Storage service was not initialized.\n")
		return fmt.Errorf("Storage service was not initialized.\n")
	}

	// Build list of exisitng files in destination bucket.
	existingFilenames, err := GetFileNamesWithPrefix(unembargoService, destBucket, prefixFileName)
	if err != nil {
		return err
	}

	// Copy files.
	pageToken := ""
	for {
		// Get list all objects in source bucket.
		sourceFiles := unembargoService.Objects.List(sourceBucket)
		sourceFiles.Prefix(prefixFileName)
		if pageToken != "" {
			sourceFiles.PageToken(pageToken)
		}
		sourceFilesList, err := sourceFiles.Context(context.Background()).Do()
		if err != nil {
			log.Printf("Objects List of source bucket failed: %v\n", err)
			return err
		}
		for _, oneItem := range sourceFilesList.Items {
			if existingFilenames[oneItem.Name] {
				// Delete the exisitng file in destBucket.
				result := unembargoService.Objects.Delete(destBucket, oneItem.Name).Do()
				if result != nil {
					log.Printf("Objects deletion from public bucket failed.\n")
					return fmt.Errorf("Objects deletion from public bucket failed.\n")
				}
			}

			if fileContent, err := unembargoService.Objects.Get(sourceBucket, oneItem.Name).Download(); err == nil {
				// Insert the object into destination bucket.
				object := &storage.Object{Name: oneItem.Name}
				_, err := unembargoService.Objects.Insert(destBucket, object).Media(fileContent.Body).Do()
				if err != nil {
					log.Printf("Objects insert failed: %v\n", err)
					return err
				}
			}
			// Delete the file in private bucket
			result := unembargoService.Objects.Delete(sourceBucket, oneItem.Name).Do()
			if result != nil {
				log.Printf("Objects deletion from private bucket failed.\n")
				return fmt.Errorf("Objects deletion from private bucket failed.\n")
			}
		}
		pageToken = sourceFilesList.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return nil
}

// The input date is integer in format yyyymmdd
// TODO: add validity check for input date.
func Unembargo(date int) error {
	f, err := os.OpenFile("UnembargoLogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return err
	}
	defer f.Close()

	log.SetOutput(f)

	if CheckWhetherUnembargo(date) {
		date_str := strconv.Itoa(date)
		input_dir := "sidestream/" + date_str[0:4] + "/" + date_str[4:6] + "/" + date_str[6:8]
		return UnEmbargoOneDayLegacyFiles(privateBucket, publicBucket, input_dir)
	}
	return fmt.Errorf("Date is too new, not qualified for unembargo.")
}
