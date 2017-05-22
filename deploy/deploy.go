package main

import (
        "fmt"
        "log"
  	"net/http"
        "github.com/m-lab/etl-embargo"
)


// For now, we can handle data for one day or a single file.
// The input URL is like: "hostname:port/submit?date=yyyymmdd&file=sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz"
func EmbargoHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query()["date"]
	filename := r.URL.Query()["file"]
	if len(date) == 0 && len(filename) == 0 {
		fmt.Fprint(w, "No date or filename there\n")
		http.NotFound(w, r)
		return
	}

        testConfig := embargo.NewEmbargoConfig("scraper-mlab-staging", "mlab-embargoed-data", "embargo-output", "")
	if filename[0] != "" {
		testConfig.EmbargoSingleFile(filename[0])
		fmt.Fprint(w, "Done with embargo single file "+filename[0]+" \n")
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
