// Package embargo implemented site IP loading from public URL or local file and check whether an IP is
// in the whitelist which is the list of all sites exceot the samknows sites.
package embargo

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// WhitelistChecker is a struct containing map EmbargoWhiteList which is the list
// of M-Lab site IP EXCEPT the Samknows sites.
type WhitelistChecker struct {
	EmbargoWhiteList map[string]struct{}
}

// GetDayOfWeek returns "Tuesday" for date "2017/05/16"
// for input filepath string like
// "sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz"
// This func is used in calculate monitoring metrics.
func GetDayOfWeek(filename string) (string, error) {
	if len(filename) < 21 {
		return "", errors.New("invalid filename")
	}
	date := filename[11:21]
	dateStr := strings.Replace(date, "/", "-", -1) + " 00:00:00"
	parsedDate, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return "", err
	}
	return parsedDate.Weekday().String(), nil
}

// FormatDateAsInt return a date in interger as format yyyymmdd.
func FormatDateAsInt(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// SiteIPURLTest is public URL of the test json file for site info.
const SiteIPURLTest = "https://storage.googleapis.com/operator-mlab-staging/metadata/v0/current/mlab-host-ips.json"

// SiteIPURL is public URL of the prod json file for site info.
const SiteIPURL = "https://storage.googleapis.com/operator-mlab-oti/metadata/v0/current/mlab-host-ips.json"

// Site is a struct for parsing json file.
type Site struct {
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Ipv6     string `json:"ipv6"`
}

// FilterSiteIPs parses bytes and returns array of struct with site IPs
// filtering out all samknows sites.
// TODO: make the filter use positive checks, including the list of things
// other than samknows, rather than excluding samknows.
func FilterSiteIPs(body []byte) (map[string]struct{}, error) {
	sites := make([]Site, 0)
	filteredIPList := make(map[string]struct{})
	if err := json.Unmarshal(body, &sites); err != nil {
		log.Printf("Cannot parse site IP json files.")
		return nil, errors.New("cannot parse site IP json files")
	}

	for _, site := range sites {
		if strings.Contains(site.Hostname, "samknows") {
			continue
		}
		if site.Ipv4 != "" {
			filteredIPList[site.Ipv4] = struct{}{}
		}
		if site.Ipv6 != "" {
			filteredIPList[site.Ipv6] = struct{}{}
		}
	}
	log.Printf("Load whitelist with length %d", len(filteredIPList))
	return filteredIPList, nil
}

// LoadFromGCS loads the embargo IP whitelist from public URL.
// TODO: add unittest for this func.
func (wc *WhitelistChecker) LoadFromGCS() error {
	project := os.Getenv("GCLOUD_PROJECT")
	log.Printf("Using project: %s\n", project)
	jsonURL := SiteIPURLTest
	if project == "mlab-oti" {
		jsonURL = SiteIPURL
	}
	log.Printf("json file used: %s", jsonURL)
	resp, err := http.Get(jsonURL)
	if err != nil {
		log.Printf("cannot download site IP json file.\n")
		return err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Cannot read site IP json files.\n")
		return err
	}

	wc.EmbargoWhiteList, err = FilterSiteIPs(body)
	return err
}

// LoadFromLocalWhitelist loads embargo IP whitelist from a local file.
func (wc *WhitelistChecker) LoadFromLocalWhitelist(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	whiteList := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		oneLine := strings.TrimSuffix(scanner.Text(), "\n")
		whiteList[oneLine] = struct{}{}
	}
	wc.EmbargoWhiteList = whiteList
	return nil
}

// CheckInWhiteList checks whether the IP in fileName is in the embargo whitelist.
// The filename is like: 20170225T23:00:00Z_4.34.58.34_0.web100
// file with IP that is in the site IP list, return true
// file with IP not in the site IP list, return false
func (wc *WhitelistChecker) CheckInWhiteList(fileName string) bool {
	fn := FileName{Name: fileName}
	localIP := fn.GetLocalIP()
	_, ok := wc.EmbargoWhiteList[localIP]
	return ok
}
