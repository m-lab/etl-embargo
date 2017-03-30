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
	embargoService = CreateService()
	sourceBucket   = ""
	destBucket     = ""
        embargoCheck   = new(EmbargoCheck)
        embargoDate    = ""
)

// EmbargoOneTar process one tar file, split it to 2 files, the embargoed files
// will be saved in a private dir, and the unembargoed part will be save in a
// public dir.
func EmbargoOneTar(content io.Reader, tarfileName string) bool {
        embargoCheck.ReadWhitelistFromGCS("whitelist")
        embargoCheck.embargodate = embargoDate
	// Create tar reader
	zipReader, err := gzip.NewReader(content)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer zipReader.Close()
	unzippedImage, err := ioutil.ReadAll(zipReader)
	if err != nil {
		fmt.Println(err)
		return false
	}
	unzippedReader := bytes.NewReader(unzippedImage)
	tarReader := tar.NewReader(unzippedReader)

	// Create buffer for output
	var privateBuf bytes.Buffer
	var publicBuf bytes.Buffer
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
			return false
		}
		//fmt.Printf(header.Name + "\n")
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
				return false
			}
			if _, err := privateTw.Write([]byte(output)); err != nil {
				fmt.Println(err)
				return false
			}
		} else {
			// put this file to a public buffer
			if err := publicTw.WriteHeader(hdr); err != nil {
				fmt.Println(err)
			}
			if _, err := publicTw.Write([]byte(output)); err != nil {
				fmt.Println(err)
				return false
			}
		}
	}

	if err := publicTw.Close(); err != nil {
		fmt.Println(err)
		return false
	}
	if err := privateTw.Close(); err != nil {
		fmt.Println(err)
		return false
	}
	if err := publicGzw.Close(); err != nil {
		fmt.Println(err)
		return false
	}
	if err := privateGzw.Close(); err != nil {
		fmt.Println(err)
		return false
	}

	publicObject := &storage.Object{Name: "public/" + tarfileName}
	privateObject := &storage.Object{Name: "private/" + tarfileName}
	if _, err := embargoService.Objects.Insert(destBucket, publicObject).Media(&publicBuf).Do(); err != nil {
		fmt.Printf("Objects insert failed: %v\n", err)
		return false
	}

	if _, err := embargoService.Objects.Insert(destBucket, privateObject).Media(&privateBuf).Do(); err != nil {
		fmt.Printf("Objects insert failed: %v\n", err)
		return false
	}
	return true
}

// Embargo do embargo ckecking to all files in the sourceBucket.
func Embargo() bool {
	if embargoService == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return false
	}

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
		if !EmbargoOneTar(fileContent.Body, oneItem.Name) {
			return false
		}
	}
	return true
}
