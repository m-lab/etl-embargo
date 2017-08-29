package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

        "github.com/m-lab/etl/storage"
        "github.com/m-lab/etl-embargo"
)

// For now, we can handle data for one day or a single file.
// TODO(dev): make sure only authorized users can call this.
// The input URL is like: "hostname:port/submit?date=yyyymmdd&file=gs://scraper-mlab-sandbox/sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz&&publicBucket=archive-mlab-sandbox&&privateBucket=embargo-mlab-sandbox"
func EmbargoHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query()["date"]
	filename := r.URL.Query()["file"]
	publicBucket := r.URL.Query()["publicBucket"]
	privateBucket := r.URL.Query()["privateBucket"]
	if len(date) == 0 && len(filename) == 0 {
		fmt.Fprint(w, "Missing date or filename there\n")
		http.NotFound(w, r)
		return
	}
	if len(publicBucket) == 0 || len(privateBucket) == 0 {
		fmt.Fprint(w, "Missing destination bucket there\n")
		http.NotFound(w, r)
		return
	}
	fn, err := storage.GetFilename(filename[0])
	if err != nil {
		log.Printf("Invalid filename: %s\n", fn)
		return
	}

	//log.Printf("filename: %s\n", fn)
	removePrefix := fn[5:]
	bucketNameEnd := strings.IndexByte(removePrefix, '/')
	sourceBucket := removePrefix[0:bucketNameEnd]
	filePath := removePrefix[bucketNameEnd+1:]

	testConfig := embargo.NewEmbargoConfig(sourceBucket, privateBucket[0], publicBucket[0], "")
	if fn != "" {
		testConfig.EmbargoSingleFile(filePath)
		fmt.Fprint(w, "Done with embargo single file "+fn+" \n")
		return
	}

	testConfig.EmbargoOneDayData(date[0])
	fmt.Fprint(w, "Done with embargo on new coming data for date: "+date[0]+" \n")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func main() {
	http.HandleFunc("/submit", EmbargoHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
