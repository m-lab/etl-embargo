// Embargo implementation.
package embargo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/context"
	storage "google.golang.org/api/storage/v1"

	"github.com/m-lab/etl-embargo/metrics"
)

type EmbargoConfig struct {
	sourceBucket      string
	destPrivateBucket string
	destPublicBucket  string
	embargoCheck      EmbargoCheck
	embargoService    *storage.Service
}

func NewEmbargoConfig(sourceBucketName, privateBucketName, publicBucketName, whitelistFile string) *EmbargoConfig {
	nc := &EmbargoConfig{
		sourceBucket:      sourceBucketName,
		destPrivateBucket: privateBucketName,
		destPublicBucket:  publicBucketName,
	}
	if whitelistFile == "" {
		if !nc.embargoCheck.LoadWhitelist() {
			log.Printf("Cannot load whitelist from GCS.\n")
			return nil
		}
	} else {
		if !nc.embargoCheck.ReadWhitelistFromLocal(whitelistFile) {
			log.Printf("Cannot load whitelist from local.\n")
			return nil
		}
	}
	nc.embargoService = CreateService()

	if nc.embargoService == nil {
		log.Fatal("Cannot create storage service.\n")
	}
	return nc
}

// Write results to GCS.
func (ec *EmbargoConfig) WriteResults(tarfileName string, embargoBuf, publicBuf bytes.Buffer) error {
	embargoTarfileName := strings.Replace(tarfileName, ".tgz", "-e.tgz", -1)
	publicObject := &storage.Object{Name: tarfileName}
	embargoObject := &storage.Object{Name: embargoTarfileName}
	if _, err := ec.embargoService.Objects.Insert(ec.destPublicBucket, publicObject).Media(&publicBuf).Do(); err != nil {
		log.Printf("Objects insert failed: %v\n", err)
		return err
	}

	if _, err := ec.embargoService.Objects.Insert(ec.destPrivateBucket, embargoObject).Media(&embargoBuf).Do(); err != nil {
		log.Printf("Objects insert failed: %v\n", err)
		return err
	}
	return nil
}

// Split one tar files into 2 buffers.
func (ec *EmbargoConfig) SplitFile(content io.Reader) (bytes.Buffer, bytes.Buffer, error) {
	var embargoBuf bytes.Buffer
	var publicBuf bytes.Buffer
	// Create tar reader
	zipReader, err := gzip.NewReader(content)
	if err != nil {
		log.Printf("zip reader failed to be created: %v\n", err)
		return embargoBuf, publicBuf, err
	}
	defer zipReader.Close()
	unzippedBytes, err := ioutil.ReadAll(zipReader)
	if err != nil {
		log.Printf("cannot read the bytes from zip reader: %v\n", err)
		return embargoBuf, publicBuf, err
	}
	unzippedReader := bytes.NewReader(unzippedBytes)
	tarReader := tar.NewReader(unzippedReader)

	embargoGzw := gzip.NewWriter(&embargoBuf)
	publicGzw := gzip.NewWriter(&publicBuf)
	embargoTw := tar.NewWriter(embargoGzw)
	publicTw := tar.NewWriter(publicGzw)

	// Handle the small files inside one tar file.
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("can not read the header file correctly: %v\n", err)
			return embargoBuf, publicBuf, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		basename := filepath.Base(header.Name)
		info := header.FileInfo()
		hdr := new(tar.Header)
		hdr.Name = header.Name
		hdr.Size = info.Size()
		hdr.Mode = int64(info.Mode())
		hdr.ModTime = info.ModTime()
		hdr.Typeflag = tar.TypeReg
		output, err := ioutil.ReadAll(tarReader)
		if ec.embargoCheck.ShouldEmbargo(basename) {
			// put this file to a private buffer
			if err := embargoTw.WriteHeader(hdr); err != nil {
				log.Printf("cannot write the embargoed header: %v\n", err)
				return embargoBuf, publicBuf, err
			}
			//log.Printf("embargo file: %s\n", basename)
			if _, err := embargoTw.Write([]byte(output)); err != nil {
				log.Printf("cannot write the embargoed content to a buffer: %v\n", err)
				return embargoBuf, publicBuf, err
			}
		} else {
			// put this file to a public buffer
			if err := publicTw.WriteHeader(hdr); err != nil {
				log.Printf("cannot write the public header: %v\n", err)
				return embargoBuf, publicBuf, err
			}
			//log.Printf("publish file: %s\n", basename)
			if _, err := publicTw.Write([]byte(output)); err != nil {
				log.Printf("cannot write the public content to a buffer: %v\n", err)
				return embargoBuf, publicBuf, err
			}
		}
	}

	if err := publicTw.Close(); err != nil {
		log.Println("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	if err := embargoTw.Close(); err != nil {
		log.Println("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	if err := publicGzw.Close(); err != nil {
		log.Println("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	if err := embargoGzw.Close(); err != nil {
		log.Println("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	return embargoBuf, publicBuf, nil
}

// EmbargoOneTar process a filepath string like
// "sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz",
// return "Tuesday" for date "2017/05/16"
func GetDayOfWeek(filename string) (string, error) {
	if len(filename) < 21 {
		return "", errors.New("invalid filename.")
	}
	date := filename[11:21]
	dateStr := strings.Replace(date, "/", "-", -1) + " 00:00:00"
	parsedDate, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return "", err
	}
	return parsedDate.Weekday().String(), nil
}

// EmbargoOneTar processes one tar file, splits it to 2 files. The embargoed files
// will be saved in a private bucket, and the unembargoed part will be save in a
// public bucket.
// The private file will have a different name, so it can be copied to public
// bucket directly when it becomes one year old.
func (ec *EmbargoConfig) EmbargoOneTar(content io.Reader, tarfileName string) error {
	dayOfWeek, err := GetDayOfWeek(tarfileName)
	if err != nil {
		metrics.Metrics_embargoErrorTotal.WithLabelValues("sidestream", "Unknown").Inc()
	}
	embargoBuf, publicBuf, err := ec.SplitFile(content)
	if err != nil {
		metrics.Metrics_embargoErrorTotal.WithLabelValues("sidestream", dayOfWeek).Inc()
		return err
	}
	if err = ec.WriteResults(tarfileName, embargoBuf, publicBuf); err != nil {
		metrics.Metrics_embargoErrorTotal.WithLabelValues("sidestream", dayOfWeek).Inc()
		return err
	}

	metrics.Metrics_embargoSuccessTotal.WithLabelValues("sidestream", dayOfWeek).Inc()
	return nil
}

// Embargo do embargo ckecking to all files in the sourceBucket.
// The input date is in format yyyy/mm/dd
// TODO: handle midway crash. Since the source bucket is unchanged, if it failed
// in the middle, we just rerun it for that specific day.
func (ec *EmbargoConfig) EmbargoOneDayData(date string) error {
	f, err := os.OpenFile("EmbargoLogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return err
	}
	defer f.Close()

	log.SetOutput(f)

	// TODO: Create service in a Singleton object, and reuse them for all GCS requests.

	if ec.embargoService == nil {
		log.Printf("Storage service was not initialized.\n")
		return fmt.Errorf("Storage service was not initialized.\n")
	}

	sourceFiles := ec.embargoService.Objects.List(ec.sourceBucket)
	sourceFiles.Prefix("sidestream/" + date)
	sourceFilesList, err := sourceFiles.Context(context.Background()).Do()
	if err != nil {
		log.Printf("Objects List of source bucket failed: %v\n", err)
		return err
	}
	for _, oneItem := range sourceFilesList.Items {
		//fmt.Printf(oneItem.Name + "\n")
		if !strings.Contains(oneItem.Name, "tgz") || !strings.Contains(oneItem.Name, "sidestream") {
			continue
		}

		fileContent, err := ec.embargoService.Objects.Get(ec.sourceBucket, oneItem.Name).Download()
		if err != nil {
			log.Printf("fail to read a tar file from the bucket: %v\n", err)
			return err
		}
		if err := ec.EmbargoOneTar(fileContent.Body, oneItem.Name); err != nil {
			return err
		}
	}
	return nil
}

func (ec *EmbargoConfig) EmbargoSingleFile(filename string) error {
	ec.embargoCheck.ReadWhitelistFromGCS(WhiteListBucket, "whitelist_full")
	if !strings.Contains(filename, "tgz") || !strings.Contains(filename, "sidestream") {
		return errors.New("Not a proper sidestream file.")
	}

	fileContent, err := ec.embargoService.Objects.Get(ec.sourceBucket, filename).Download()
	if err != nil {
		log.Printf("fail to read a tar file from the bucket: %v\n", err)
		return err
	}
	if err := ec.EmbargoOneTar(fileContent.Body, filename); err != nil {
		return err
	}
	return nil
}
