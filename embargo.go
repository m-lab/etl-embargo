// Embargo implementation.
package embargo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
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
func WriteResults(tarfileName string, service *storage.Service, privateBuf, publicBuf bytes.Buffer) bool {
	privateTarfileName := strings.Replace(tarfileName, ".tgz", "-e.tgz", -1)
	publicObject := &storage.Object{Name: tarfileName}
	privateObject := &storage.Object{Name: privateTarfileName}
	if _, err := service.Objects.Insert(destPublicBucket, publicObject).Media(&publicBuf).Do(); err != nil {
		fmt.Printf("Objects insert failed: %v\n", err)
		return false
	}

	if _, err := service.Objects.Insert(destPrivateBucket, privateObject).Media(&privateBuf).Do(); err != nil {
		fmt.Printf("Objects insert failed: %v\n", err)
		return false
	}
	return true
}

// Split one tar files into 2 buffers.
func embargoBuf(content io.Reader) (bytes.Buffer, bytes.Buffer, error) {
	var privateBuf bytes.Buffer
	var publicBuf bytes.Buffer
	// Create tar reader
	zipReader, err := gzip.NewReader(content)
	if err != nil {
		fmt.Println(err)
		return privateBuf, publicBuf, err
	}
	defer zipReader.Close()
	unzippedImage, err := ioutil.ReadAll(zipReader)
	if err != nil {
		fmt.Println(err)
		return privateBuf, publicBuf, err
	}
	unzippedReader := bytes.NewReader(unzippedImage)
	tarReader := tar.NewReader(unzippedReader)

	privateGzw := gzip.NewWriter(&privateBuf)
	publicGzw := gzip.NewWriter(&publicBuf)
	privateTw := tar.NewWriter(privateGzw)
	publicTw := tar.NewWriter(publicGzw)

	// Handle one tar file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return privateBuf, publicBuf, err
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
			if err := privateTw.WriteHeader(hdr); err != nil {
				fmt.Println(err)
				return privateBuf, publicBuf, err
			}
			if _, err := privateTw.Write([]byte(output)); err != nil {
				fmt.Println(err)
				return privateBuf, publicBuf, err
			}
		} else {
			// put this file to a public buffer
			if err := publicTw.WriteHeader(hdr); err != nil {
				fmt.Println(err)
			}
			if _, err := publicTw.Write([]byte(output)); err != nil {
				fmt.Println(err)
				return privateBuf, publicBuf, err
			}
		}
	}

	if err := publicTw.Close(); err != nil {
		fmt.Println(err)
		return privateBuf, publicBuf, err
	}
	if err := privateTw.Close(); err != nil {
		fmt.Println(err)
		return privateBuf, publicBuf, err
	}
	if err := publicGzw.Close(); err != nil {
		fmt.Println(err)
		return privateBuf, publicBuf, err
	}
	if err := privateGzw.Close(); err != nil {
		fmt.Println(err)
		return privateBuf, publicBuf, err
	}
	return privateBuf, publicBuf, nil
}

// EmbargoOneTar process one tar file, split it to 2 files, the embargoed files
// will be saved in a private dir, and the unembargoed part will be save in a
// public dir.
// The private file will have a different name, so it can be copied to public
// bucket directly when it becomes one year old.
func EmbargoOneTar(content io.Reader, tarfileName string, service *storage.Service) bool {
	privateBuf, publicBuf, err := embargoBuf(content)
	if err == nil && WriteResults(tarfileName, service, privateBuf, publicBuf) {
		return true
	}
	return false
}

// Embargo do embargo ckecking to all files in the sourceBucket.
func Embargo() bool {
	embargoService := CreateService()
	if embargoService == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}
	embargoCheck.ReadWhitelistFromGCS("whitelist")
	embargoCheck.Embargodate = embargoDate
	sourceFiles := embargoService.Objects.List(sourceBucket)
	sourceFilesList, err := sourceFiles.Context(context.Background()).Do()
	if err != nil {
		fmt.Printf("Objects List of source bucket failed: %v\n", err)
		return false
	}
	for _, oneItem := range sourceFilesList.Items {
		//fmt.Printf(oneItem.Name + "\n")
		if !strings.Contains(oneItem.Name, "tgz") || !strings.Contains(oneItem.Name, "sidestream") {
			continue
		}

		fileContent, err := embargoService.Objects.Get(sourceBucket, oneItem.Name).Download()
		if err != nil {
			fmt.Println(err)
			return false
		}
		if !EmbargoOneTar(fileContent.Body, oneItem.Name, embargoService) {
			return false
		}
	}
	return true
}
