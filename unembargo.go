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
	"errors"
	"fmt"
	"golang.org/x/net/context"
	storage_v1 "google.golang.org/api/storage/v1"
	"log"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"

	"github.com/m-lab/etl-embargo/metrics"
)

type UnembargoConfig struct {
	privateBucket string
	publicBucket  string
}

func NewUnembargoConfig(privateBucketName, publicBucketName string) *UnembargoConfig {
	nc := &UnembargoConfig{
		privateBucket: privateBucketName,
		publicBucket:  publicBucketName,
	}
	return nc
}

// Get filenames for given bucket with the given prefix. Use the service
func GetFileNamesWithPrefix(service *storage_v1.Service, bucketName string, prefixFileName string) (map[string]bool, error) {
	existingFilenames := make(map[string]bool)
	pageToken := ""
	for {
		destinationFiles := service.Objects.List(bucketName)

		destinationFiles.Prefix(prefixFileName)
		destinationFiles.PageToken(pageToken)
		destinationFilesList, err := destinationFiles.Context(context.Background()).Do()
		if err != nil {
			log.Printf("Objects.List failed: %v\n", err)
			return existingFilenames, err
		}
		for _, oneItem := range destinationFilesList.Items {
			existingFilenames[oneItem.Name] = true
		}
		pageToken = destinationFilesList.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return existingFilenames, nil
}

// UnEmbargoOneDayLegacyFiles unembargos one day data in the sourceBucket,
// and writes the output to destBucket.
// The date is used as prefixFileName in format sidestream/yyyy/mm/dd
func UnEmbargoOneDayLegacyFiles(sourceBucket string, destBucket string, prefixFileName string) error {
	unembargoService := CreateService()
	if unembargoService == nil {
		log.Printf("Storage service was not initialized.\n")
		return fmt.Errorf("Storage service was not initialized.\n")
	}
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return err
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
		sourceFiles.PageToken(pageToken)
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
			// Copy the file to dest bucket.
			// CopierFrom() is only available in newer "cloud.google.com/go/storage" libraty
			src := client.Bucket(sourceBucket).Object(oneItem.Name)
			dst := client.Bucket(destBucket).Object(oneItem.Name)
			if _, err := dst.CopierFrom(src).Run(context.Background()); err != nil {
				return fmt.Errorf("Objects copy failed: %v\n", err)
			}
			// Do not delete the file in private bucket
			//result := unembargoService.Objects.Delete(sourceBucket, oneItem.Name).Do()
			//if result != nil {
			//	log.Printf("Objects deletion from private bucket failed.\n")
			//	return fmt.Errorf("Objects deletion from private bucket failed.\n")
			//}
			metrics.Metrics_unembargoTarTotal.WithLabelValues("sidestream").Inc()
		}
		pageToken = sourceFilesList.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return nil
}

// Unembargo unembargo the data of the input date in format yyyymmdd.
// TODO(dev): add more validity check for input date.
func (nc *UnembargoConfig) Unembargo(date int) error {
	if date <= 20160000 || date > 21000000 {
		return errors.New("The date is out of range.")
	}

	f, err := os.OpenFile("UnembargoLogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return err
	}
	defer f.Close()

	log.SetOutput(f)

	if date <= FormatDateAsInt(time.Now().AddDate(-1, 0, 0)) {
		dateStr := strconv.Itoa(date)
		inputDir := "sidestream/" + dateStr[0:4] + "/" + dateStr[4:6] + "/" + dateStr[6:8]
		return UnEmbargoOneDayLegacyFiles(nc.privateBucket, nc.publicBucket, inputDir)
	}
        log.Printf("Date is too new, not qualified for unembargo.")
	return fmt.Errorf("Date is too new, not qualified for unembargo.")
}

func UnembargoCron(date int) error {
	project := os.Getenv("GCLOUD_PROJECT")
	log.Printf("current project: %s", project)
	privateBucketName := "embargo-" + project
	publicBucketName := "archive-" + project

	uc := NewUnembargoConfig(privateBucketName, publicBucketName)
	return uc.Unembargo(date)
}
