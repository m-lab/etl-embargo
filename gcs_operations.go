// Package gcs implements a simple library for basic operations given bucket
// names and file name/prefix, such as ls, cp, rm, etc. on Google Cloud
// Storage.
package embargo

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
)

// CreateService creates GCS service used by the following functions.
func CreateService() *storage.Service {
	// This scope allows the application full control over resources in Google Cloud Storage
	var scope = storage.DevstorageFullControlScope
	client, err := google.DefaultClient(context.Background(), scope)
	if err != nil {
		fmt.Printf("Unable to get default storage client: %v \n", err)
		return nil
	}
	service, err := storage.New(client)
	if err != nil {
		fmt.Printf("Unable to create storage service: %v\n", err)
		return nil
	}
	return service
}

// TODO: Create service in a Singleton object, and reuse them for all GCS requests.

// CreateBucket creates a new bucket. Return true if it already exsits or is created successfully.
func CreateBucket(projectID string, bucketName string) bool {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

	if _, err := service.Buckets.Get(bucketName).Do(); err == nil {
		fmt.Printf("Bucket %s already exsits.\n", bucketName)
		return true
	} else {
		// Create a bucket.
		if res, err := service.Buckets.Insert(projectID, &storage.Bucket{Name: bucketName}).Do(); err == nil {
			fmt.Printf("Created bucket %v at location %v\n", res.Name, res.SelfLink)
		} else {
			fmt.Printf("Failed creating bucket %s: %v\n", bucketName, err)
		}
	}
	return true
}

// GetFileNamesFromBucket returns array of file names in that bucket given the bucket name,. ("ls")
func GetFileNamesFromBucket(bucketName string) []string {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return nil
	}

	var fileNames []string
	pageToken := ""
	for {
		call := service.Objects.List(bucketName)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Do()
		if err != nil {
			fmt.Printf("Get file list failed: %v\n", err)
			return nil
		}
		for _, object := range res.Items {
			fileNames = append(fileNames, object.Name)
		}
		if pageToken = res.NextPageToken; pageToken == "" {
			break
		}
	}
	return fileNames
}

// DeleteFiles deletes all files with specified prefix from bucket. ("rm")
func DeleteFiles(bucketName string, prefixFileName string) bool {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

	_, err := service.Buckets.Get(bucketName).Do()
	if err != nil {
		fmt.Printf("Bucket %s does not exists.\n", bucketName)
		return false
	}

	// Delete files.
	pageToken := ""
	for {
		// Get list all objects in source bucket.
		sourceFiles := service.Objects.List(bucketName)
		sourceFiles.Prefix(prefixFileName)
		if pageToken != "" {
			sourceFiles.PageToken(pageToken)
		}
		sourceFilesList, err := sourceFiles.Context(context.Background()).Do()
		if err != nil {
			fmt.Printf("Objects List of source bucket failed: %v\n", err)
			return false
		}
		for _, oneItem := range sourceFilesList.Items {
			result := service.Objects.Delete(bucketName, oneItem.Name).Do()
			if result != nil {
				fmt.Printf("Objects deletion failed: %v\n", err)
				return false
			}
		}

		if pageToken = sourceFilesList.NextPageToken; pageToken == "" {
			break
		}
	}
	return true
}

// Delete the bucket if it is empty. ("rmdir")
func DeleteBucket(bucketName string) bool {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

	sourceFiles, err := service.Objects.List(bucketName).Do()
	if err != nil {
		return false
	}
	if len(sourceFiles.Items) == 0 {
		if err := service.Buckets.Delete(bucketName).Do(); err != nil {
			fmt.Printf("Could not delete bucket %v\n", err)
			return false
		} else {
			fmt.Printf("Delete bucket %s successfully.\n", bucketName)
			return true
		}
	} else {
		fmt.Printf("Could not delete non empty bucket %v\n", bucketName)
		return false
	}

}

// UploadFile uploads one file from local path to bucket. ("cp")
func UploadFile(bucketName string, fileName string, targetdir string) bool {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening local file %s: %v\n", fileName, err)
		return false
	}
	object := &storage.Object{Name: targetdir + filepath.Base(fileName)}
	if res, err := service.Objects.Insert(bucketName, object).Media(file).Do(); err == nil {
		fmt.Printf("Created object %v at location %v\n", res.Name, res.SelfLink)
		return true
	}
	fmt.Printf("Objects.Insert failed: %v\n", err)
	return false
}

// CopyOneFile copies one file from one bucket to another bucket. Return true if succeed. ("cp")
func CopyOneFile(sourceBucket string, destBucket string, fileName string) bool {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

	if fileContent, err := service.Objects.Get(sourceBucket, fileName).Download(); err == nil {
		object := &storage.Object{Name: fileName}
		_, err := service.Objects.Insert(destBucket, object).Media(fileContent.Body).Do()
		if err != nil {
			fmt.Printf("Objects insert failed: %v\n", err)
			return false
		}
	}
	return true
}

// SyncTwoBuckets copies all files with PrefixFileName from SourceBucke to DestBucket if there
// is no one yet. Return true if succeed.
func SyncTwoBuckets(sourceBucket string, destBucket string, prefixFileName string) bool {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

	// Build list of exisitng files in destination bucket.
	existingFilenames := make(map[string]bool)
	destPageToken := ""
	for {
		destinationFiles := service.Objects.List(destBucket)
		if destPageToken != "" {
			destinationFiles.PageToken(destPageToken)
		}
		destinationFiles.Prefix(prefixFileName)
		destinationFilesList, err := destinationFiles.Context(context.Background()).Do()
		if err != nil {
			fmt.Printf("Objects.List failed: %v\n", err)
			return false
		}
		for _, oneItem := range destinationFilesList.Items {
			existingFilenames[oneItem.Name] = true
		}
		destPageToken = destinationFilesList.NextPageToken
		if destPageToken == "" {
			break
		}
	}

	// Copy files.
	pageToken := ""
	for {
		// Get list all objects in source bucket.
		sourceFiles := service.Objects.List(sourceBucket)
		sourceFiles.Prefix(prefixFileName)
		if pageToken != "" {
			sourceFiles.PageToken(pageToken)
		}
		sourceFilesList, err := sourceFiles.Context(context.Background()).Do()
		if err != nil {
			fmt.Printf("Objects List of source bucket failed: %v\n", err)
			return false
		}
		for _, oneItem := range sourceFilesList.Items {
			if existingFilenames[oneItem.Name] {
				fmt.Printf("object %s already there\n", oneItem.Name)
				continue
			}
			if fileContent, err := service.Objects.Get(sourceBucket, oneItem.Name).Download(); err == nil {
				// Insert the object into destination bucket.
				object := &storage.Object{Name: oneItem.Name}
				_, err := service.Objects.Insert(destBucket, object).Media(fileContent.Body).Do()
				if err != nil {
					fmt.Printf("Objects insert failed: %v\n", err)
					return false
				}
			}
		}
		pageToken = sourceFilesList.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return true
}

// CompareBuckets compares whether 2 buckets have exactly same files. Return true if they are the same.
func CompareBuckets(sourceBucket string, destBucket string) bool {
	service := CreateService()
	if service == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

	// Build list of exisitng files in destination bucket.
	existingFilenames := make(map[string]bool)
	destPageToken := ""
	for {
		destinationFiles := service.Objects.List(destBucket)
		if destPageToken != "" {
			destinationFiles.PageToken(destPageToken)
		}
		destinationFilesList, err := destinationFiles.Context(context.Background()).Do()
		if err != nil {
			fmt.Printf("Objects.List failed: %v\n", err)
			return false
		}
		for _, oneItem := range destinationFilesList.Items {
			existingFilenames[oneItem.Name] = true
		}
		destPageToken = destinationFilesList.NextPageToken
		if destPageToken == "" {
			break
		}
	}

	// Go through files in destination.
	pageToken := ""
	for {
		// Get list all objects in source bucket.
		sourceFiles := service.Objects.List(sourceBucket)
		if pageToken != "" {
			sourceFiles.PageToken(pageToken)
		}
		sourceFilesList, err := sourceFiles.Context(context.Background()).Do()
		if err != nil {
			fmt.Printf("Objects List of source bucket failed: %v\n", err)
			return false
		}
		for _, oneItem := range sourceFilesList.Items {
			if existingFilenames[oneItem.Name] {
				// File is already there. We need to compare the size.
				existingFilenames[oneItem.Name] = false
				continue
			} else {
				fmt.Printf("Here is a file in sourceBucket but not in destBucket: %s", oneItem.Name)
				return false
			}

		}
		pageToken = sourceFilesList.NextPageToken
		if pageToken == "" {
			break
		}
	}

	// Go through map existingFilenames[] to see whether all of them are flipped to false.
	for fileName := range existingFilenames {
		if existingFilenames[fileName] == true {
			fmt.Printf("Here is a file in destBucket but not in sourceBucket: %s", fileName)
			return false
		}
	}

	return true
}
