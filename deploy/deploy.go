package main

import (
	"fmt"
	"log"
	"net/http"
	// Enable exported debug vars.  See https://golang.org/pkg/expvar/
	_ "expvar"
	"strings"
	"time"

	"github.com/m-lab/etl-embargo"
	"github.com/m-lab/etl-embargo/metrics"
	"github.com/m-lab/etl/storage"
)

// EmbargoHandler handles data for one day or a single file.
// TODO(dev): make sure only authorized users can call this.
// For example, if we want to process embargo on
// gs://scraper-mlab-sandbox/sidestream/2017/05/29/20170529T000000Z-mlab1-atl02-sidestream-0000.tgz
// The input URL is like: "https://embargo-dot-mlab-sandbox.appspot.com/submit?file=Z3M6Ly9zY3JhcGVyLW1sYWItc2FuZGJveC9zaWRlc3RyZWFtLzIwMTcvMDUvMjkvMjAxNzA1MjlUMDAwMDAwWi1tbGFiMS1hdGwwMi1zaWRlc3RyZWFtLTAwMDAudGd6&&publicBucket=archive-mlab-sandbox&&privateBucket=embargo-mlab-sandbox"
func EmbargoHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query()["date"]
	filename := r.URL.Query()["file"]
	if len(date) == 0 && len(filename) == 0 {
		fmt.Fprint(w, "Missing date or filename there\n")
		http.NotFound(w, r)
		return
	}

	fn, err := storage.GetFilename(filename[0])
	if err != nil {
		log.Printf("Invalid filename: %s\n", fn)
		http.Error(w, "Invalid filename: "+fn, http.StatusInternalServerError)
		return
	}

	//log.Printf("filename: %s\n", fn)
	removePrefix := fn[5:]
	bucketNameEnd := strings.IndexByte(removePrefix, '/')
	filePath := removePrefix[bucketNameEnd+1:]

	testConfig, err := embargo.GetEmbargoConfig("")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fn != "" {
		err := testConfig.EmbargoSingleFile(filePath)
		if err != nil {
			log.Print("Fail with embargo single file " + fn + " \n")
			http.Error(w, "Fail with embargo single file.", http.StatusInternalServerError)
		}
		return
	}

	// Process the date if there is not single file.
	if len(date) > 0 {
		err := testConfig.EmbargoOneDayData(date[0], embargo.FormatDateAsInt(time.Now().AddDate(-1, 0, 0)))
		if err != nil {
			log.Print("Fail with embargo on new coming data for date: " + date[0] + " \n")
			http.Error(w, "Fail with embargo on new coming data for date: "+date[0]+" \n", http.StatusInternalServerError)
			return
		}
		log.Print("success with mebargo one file")
		return
	}
}

// Update the embargo whitelist by reloading the site IPs daily
func updateEmbargoWhitelist(w http.ResponseWriter, r *http.Request) {
	log.Printf("Update the site IPs used for embargo process.\n")

	err := embargo.UpdateWhitelist()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func main() {
	http.HandleFunc("/submit", EmbargoHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	http.HandleFunc("/cron/update_embargo_whitelist", updateEmbargoWhitelist)
	metrics.SetupPrometheus()
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
