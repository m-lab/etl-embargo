// Embargo implementation.
package embargo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
	storage "google.golang.org/api/storage/v1"
)

var (
	sourceBucket      = ""
	destPrivateBucket = ""
	destPublicBucket  = ""
	embargoCheck      = new(EmbargoCheck)
	embargoDate       = 0
)

// Write results to GCS.
func WriteResults(tarfileName string, service *storage.Service, embargoBuf, publicBuf bytes.Buffer) error {
	embargoTarfileName := strings.Replace(tarfileName, ".tgz", "-e.tgz", -1)
	publicObject := &storage.Object{Name: tarfileName}
	embargoObject := &storage.Object{Name: embargoTarfileName}
	if _, err := service.Objects.Insert(destPublicBucket, publicObject).Media(&publicBuf).Do(); err != nil {
		log.Printf("Objects insert failed: %v\n", err)
		return err
	}

	if _, err := service.Objects.Insert(destPrivateBucket, embargoObject).Media(&embargoBuf).Do(); err != nil {
		log.Printf("Objects insert failed: %v\n", err)
		return err
	}
	return nil
}

// Split one tar files into 2 buffers.
func SplitFile(content io.Reader) (bytes.Buffer, bytes.Buffer, error) {
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
		output, err := ioutil.ReadAll(tarReader)
		if strings.Contains(basename, "web100") && embargoCheck.ShouldEmbargo(basename) {
			// put this file to a private buffer
			if err := embargoTw.WriteHeader(hdr); err != nil {
				log.Printf("cannot write the embargoed header: %v\n", err)
				return embargoBuf, publicBuf, err
			}
			if _, err := embargoTw.Write([]byte(output)); err != nil {
				log.Printf("cannot write the embargoed content to a buffer: %v\n", err)
				return embargoBuf, publicBuf, err
			}
		} else {
			// put this file to a public buffer
			if err := publicTw.WriteHeader(hdr); err != nil {
				log.Printf("cannot write the public header: %v\n", err)
			}
			if _, err := publicTw.Write([]byte(output)); err != nil {
				log.Printf("cannot write the public content to a buffer: %v\n", err)
				return embargoBuf, publicBuf, err
			}
		}
	}

	if err := publicTw.Close(); err != nil {
		log.Printf("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	if err := embargoTw.Close(); err != nil {
		log.Printf("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	if err := publicGzw.Close(); err != nil {
		log.Printf("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	if err := embargoGzw.Close(); err != nil {
		log.Printf("cannot close tar writer", err)
		return embargoBuf, publicBuf, err
	}
	return embargoBuf, publicBuf, nil
}

// EmbargoOneTar processes one tar file, splits it to 2 files. The embargoed files
// will be saved in a private bucket, and the unembargoed part will be save in a
// public bucket.
// The private file will have a different name, so it can be copied to public
// bucket directly when it becomes one year old.
func EmbargoOneTar(content io.Reader, tarfileName string, service *storage.Service) error {
	embargoBuf, publicBuf, err := SplitFile(content)
	if err != nil {
		return err
	}
	if err = WriteResults(tarfileName, service, embargoBuf, publicBuf); err != nil {
		return err
	}
	return nil
}

// Embargo do embargo ckecking to all files in the sourceBucket.
// The input date is in format yyyy/mm/dd
// TODO: handle midway crash. Since the source bucket is unchanged, if it failed
// in the middle, we just rerun it for that specific day.
func EmbargoOneDayData(date string) error {
	f, err := os.OpenFile("EmbargoLogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return err
	}
	defer f.Close()

	log.SetOutput(f)

	// TODO: Create service in a Singleton object, and reuse them for all GCS requests.
	embargoService := CreateService()
	if embargoService == nil {
		log.Printf("Storage service was not initialized.\n")
		return fmt.Errorf("Storage service was not initialized.\n")
	}

	embargoCheck.ReadWhitelistFromGCS("whitelist")
	embargoCheck.Embargodate = embargoDate
	sourceFiles := embargoService.Objects.List(sourceBucket)
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

		fileContent, err := embargoService.Objects.Get(sourceBucket, oneItem.Name).Download()
		if err != nil {
			log.Printf("fail to read a tar file from the bucket: %v\n", err)
			return err
		}
		if err := EmbargoOneTar(fileContent.Body, oneItem.Name, embargoService); err != nil {
			return err
		}
	}
	return nil
}
