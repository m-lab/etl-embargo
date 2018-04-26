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

// SiteIPCheck is a struct that contains map SiteIPList is the list of M-Lab site IP EXCEPT the samknows sites.
type SiteIPCheck struct {
	SiteIPList map[string]struct{}
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
// SiteIPURLTest is public URL of the prod json file for site info.
const SiteIPURL = "https://storage.googleapis.com/operator-mlab-oti/metadata/v0/current/mlab-host-ips.json"

type Site struct {
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Ipv6     string `json:"ipv6"`
}

// IPMapFromJson parses bytes into array of struct
func IPMapFromJson(body []byte) (map[string]struct{}, error) {
	sites := make([]Site, 0)
	SiteIPList := make(map[string]struct{})
	if err := json.Unmarshal(body, &sites); err != nil {
		log.Printf("Cannot parse site IP json files.")
		return nil, errors.New("Cannot parse site IP json files.")
	}

	for _, site := range sites {
		if strings.Contains(site.Hostname, "samknows") {
			continue
		}
		if site.Ipv4 != "" {
			SiteIPList[site.Ipv4] = struct{}{}
		}
		if site.Ipv6 != "" {
			SiteIPList[site.Ipv6] = struct{}{}
		}
	}
	return SiteIPList, nil
}

// LoadSiteIPJson load the site IP json from public URL.
func (sc *SiteIPCheck) LoadSiteIPJson() error {
	project := os.Getenv("GCLOUD_PROJECT")
	log.Printf("Using project: %s\n", project)
	json_url := SiteIPURLTest
	if project == "mlab-oti" {
		json_url = SiteIPURL
	}

	resp, err := http.Get(json_url)
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

	sc.SiteIPList, err = IPMapFromJson(body)
	return err
}

// ReadSiteIPlistFromLocal loads site IP list from a local file.
func (ec *SiteIPCheck) ReadSiteIPlistFromLocal(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	siteIPList := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		oneLine := strings.TrimSuffix(scanner.Text(), "\n")
		siteIPList[oneLine] = struct{}{}
	}
	ec.SiteIPList = siteIPList
	return nil
}

// CheckInSiteIPList checks whether the IP in fileName is in the site IP list.
// The filename is like: 20170225T23:00:00Z_4.34.58.34_0.web100
// file with IP that is in the site IP list, return true
// file with IP not in the site IP list, return false
func (sc *SiteIPCheck) CheckInSiteIPList(fileName string) bool {
	fn := FileName{Name: fileName}
	localIP := fn.GetLocalIP()
	_, ok := sc.SiteIPList[localIP]
	return ok
}
