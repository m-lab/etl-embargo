// Package embargo performs embargo for all sidestream data. For all data that
// are more than one year old, or server IP in the list of M-Lab server IP list
// except the samknow sites, the sidestream test will be published.
// Otherwise the test will be embargoed and saved in a private bucket. It will
// published later when it is more than one year old.
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
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	storage "google.golang.org/api/storage/v1"

	"github.com/m-lab/etl-embargo/metrics"
)

// EmbargoConfig is a struct that performs all embargo procedures.
type EmbargoConfig struct {
	sourceBucket      string
	destPrivateBucket string
	destPublicBucket  string
	whitelistChecker  WhitelistChecker
	embargoService    *storage.Service
}

// EmbargoSingleton is the singleton object that is the pointer of the EmbargoConfig object.
var EmbargoSingleton *EmbargoConfig

// projectToURL is a map from project name to the corresponding public URL for the mlab site IP json file.
var projectToURL = map[string]string{
	"mlab-sandbox": "https://storage.googleapis.com/operator-mlab-sandbox/metadata/v0/current/mlab-host-ips.json",
	"mlab-testing": "https://storage.googleapis.com/operator-mlab-sandbox/metadata/v0/current/mlab-host-ips.json",
	"mlab-staging": "https://storage.googleapis.com/operator-mlab-staging/metadata/v0/current/mlab-host-ips.json",
	"mlab-oti":     "https://storage.googleapis.com/operator-mlab-oti/metadata/v0/current/mlab-host-ips.json",
}

func init() {
	EmbargoSingleton = nil
}

// GetEmbargoConfig creates a new EmbargoConfig and returns it.
func GetEmbargoConfig(siteIPFile string) (*EmbargoConfig, error) {
	if EmbargoSingleton != nil {
		return EmbargoSingleton, nil
	}
	project := os.Getenv("GCLOUD_PROJECT")
	log.Printf("current project: %s", project)
	ec := &EmbargoConfig{
		sourceBucket:      "scraper-" + project,
		destPrivateBucket: "embargo-" + project,
		destPublicBucket:  "archive-" + project,
	}

	jsonURL, ok := projectToURL[project]
	// The project must be one of "mlab-sandbox", "mlab-staging", "mlab-oti", or "mlab-testing".
	if !ok {
		return nil, errors.New("this job is running in wrong project")
	}
	log.Printf("json file of site IPs: %s", jsonURL)
	if siteIPFile == "" {
		err := ec.whitelistChecker.LoadFromURL(jsonURL)
		if err != nil {
			log.Printf("Cannot load site IP list from GCS.\n")
			return nil, err
		}
	} else {
		err := ec.whitelistChecker.LoadFromLocalWhitelist(siteIPFile)
		if err != nil {
			log.Printf("Cannot load site IP file from local.\n")
			return nil, err
		}
	}
	ec.embargoService = CreateService()
	if ec.embargoService == nil {
		log.Printf("Cannot create storage service.\n")
		return nil, errors.New("cannot create storage service")
	}
	EmbargoSingleton = ec
	return ec, nil
}

// UpdateWhitelist loads the site IP json file again and updates the whitelist in memory.
func UpdateWhitelist() error {
	_, err := GetEmbargoConfig("")
	if err != nil {
		return err
	}
	return nil
}

// WriteResults writes results to GCS.
func (ec *EmbargoConfig) WriteResults(tarfileName string, embargoBuf, publicBuf bytes.Buffer) error {
	embargoTarfileName := strings.Replace(tarfileName, ".tgz", "-e.tgz", -1)
	publicObject := &storage.Object{Name: tarfileName}
	embargoObject := &storage.Object{Name: embargoTarfileName}
	if _, err := ec.embargoService.Objects.Insert(ec.destPublicBucket, publicObject).Media(&publicBuf).Do(); err != nil {
		log.Printf("Objects insert failed: %v\n", err)
		return err
	} else {
		metrics.Metrics_embargoTarOutputTotal.WithLabelValues("sidestream", "public").Inc()
	}

	if _, err := ec.embargoService.Objects.Insert(ec.destPrivateBucket, embargoObject).Media(&embargoBuf).Do(); err != nil {
		log.Printf("Objects insert failed: %v\n", err)
		return err
	} else {
		metrics.Metrics_embargoTarOutputTotal.WithLabelValues("sidestream", "private").Inc()
	}
	return nil
}

// SplitFile splits one tar files into 2 buffers.
func (ec *EmbargoConfig) SplitFile(content io.Reader, moreThanOneYear bool) (bytes.Buffer, bytes.Buffer, error) {
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
		if err != nil {
			log.Printf("cannot read the tar file: %v\n", err)
			return embargoBuf, publicBuf, err
		}
		if moreThanOneYear || !strings.Contains(basename, "web100") || ec.whitelistChecker.CheckInWhiteList(basename) {
			// put this file to a public buffer
			if strings.Contains(basename, "web100") {
				metrics.Metrics_embargoFileTotal.WithLabelValues("sidestream", "public").Inc()
			}
			if err := publicTw.WriteHeader(hdr); err != nil {
				log.Printf("cannot write the public header: %v\n", err)
				return embargoBuf, publicBuf, err
			}
			if _, err := publicTw.Write([]byte(output)); err != nil {
				log.Printf("cannot write the public content to a buffer: %v\n", err)
				return embargoBuf, publicBuf, err
			}
		} else {
			// put this file to a private buffer
			if strings.Contains(basename, "web100") {
				metrics.Metrics_embargoFileTotal.WithLabelValues("sidestream", "private").Inc()
			}
			if err := embargoTw.WriteHeader(hdr); err != nil {
				log.Printf("cannot write the embargoed header: %v\n", err)
				return embargoBuf, publicBuf, err
			}
			if _, err := embargoTw.Write([]byte(output)); err != nil {
				log.Printf("cannot write the embargoed content to a buffer: %v\n", err)
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

// EmbargoOneTar processes one tar file, splits it to 2 files. The embargoed files
// will be saved in a private bucket, and the unembargoed part will be save in a
// public bucket.
// The private file will have a different name, so it can be copied to public
// bucket directly when it becomes one year old.
// The tarfileName is like 20170516T000000Z-mlab1-atl06-sidestream-0000.tgz
func (ec *EmbargoConfig) EmbargoOneTar(content io.Reader, tarfileName string, moreThanOneYear bool) error {
	embargoBuf, publicBuf, err := ec.SplitFile(content, moreThanOneYear)
	if err != nil {
		metrics.Metrics_embargoTarInputTotal.WithLabelValues("sidestream", "error").Inc()
		return err
	}
	if err = ec.WriteResults(tarfileName, embargoBuf, publicBuf); err != nil {
		metrics.Metrics_embargoTarInputTotal.WithLabelValues("sidestream", "error").Inc()
		return err
	}

	metrics.Metrics_embargoTarInputTotal.WithLabelValues("sidestream", "success").Inc()
	return nil
}

// EmbargoOneDayData do embargo for one day files.
// The input date is string in format yyyymmdd
// The cutoffDate is integer in format yyyymmdd
// TODO: handle midway crash. Since the source bucket is unchanged, if it failed
// in the middle, we just rerun it for that specific day.
func (ec *EmbargoConfig) EmbargoOneDayData(date string, cutoffDate int) error {
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
		return fmt.Errorf("storage service was not initialized")
	}

	sourceFiles := ec.embargoService.Objects.List(ec.sourceBucket)
	sourceFiles.Prefix("sidestream/" + date[0:4] + "/" + date[4:6] + "/" + date[6:8])
	sourceFilesList, err := sourceFiles.Context(context.Background()).Do()
	if err != nil {
		log.Printf("Objects List of source bucket failed: %v\n", err)
		return err
	}
	dateInteger, err := strconv.Atoi(date[0:8])
	if err != nil {
		log.Printf("Cannot get valid date: %v\n", err)
		return err
	}
	moreThanOneYear := dateInteger < cutoffDate
	for _, oneItem := range sourceFilesList.Items {
		if !strings.Contains(oneItem.Name, "tgz") || !strings.Contains(oneItem.Name, "sidestream") {
			continue
		}

		fileContent, err := ec.embargoService.Objects.Get(ec.sourceBucket, oneItem.Name).Download()
		if err != nil {
			log.Printf("fail to read a tar file from the bucket: %v\n", err)
			return err
		}
		if err := ec.EmbargoOneTar(fileContent.Body, oneItem.Name, moreThanOneYear); err != nil {
			return err
		}
	}
	return nil
}

// EmbargoSingleFile embargo the input file.
func (ec *EmbargoConfig) EmbargoSingleFile(filename string) error {
	if !strings.Contains(filename, "tgz") || !strings.Contains(filename, "sidestream") {
		return errors.New("not a proper sidestream file")
	}

	fileContent, err := ec.embargoService.Objects.Get(ec.sourceBucket, filename).Download()
	if err != nil {
		log.Printf("fail to read tar file from the bucket: %v\n", err)
		return err
	}
	baseName := filepath.Base(filename)
	dateInteger, err := strconv.Atoi(baseName[0:8])
	if err != nil {
		log.Printf("fail to get valid date from filename: %v\n", err)
		return err
	}

	moreThanOneYear := dateInteger < FormatDateAsInt(time.Now().AddDate(-1, 0, 0))

	if err := ec.EmbargoOneTar(fileContent.Body, filename, moreThanOneYear); err != nil {
		return err
	}
	return nil
}
