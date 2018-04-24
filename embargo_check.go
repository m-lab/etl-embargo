// Implement whitelist loading and embargo check based on filename.
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

type EmbargoCheck struct {
	Whitelist map[string]bool
}

// GetDayOfWeek returns "Tuesday" for date "2017/05/16"
// for input filepath string like
// "sidestream/2017/05/16/20170516T000000Z-mlab1-atl06-sidestream-0000.tgz",
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

// FormatDateAsInt return a date in interger as format yyyymmdd.
func FormatDateAsInt(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

const SITE_IP_URL_TEST = "https://storage.googleapis.com/operator-mlab-staging/metadata/v0/current/mlab-host-ips.json"
const SITE_IP_URL = "https://storage.googleapis.com/operator-mlab-oti/metadata/v0/current/mlab-host-ips.json"

type Site struct {
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Ipv6     string `json:"ipv6"`
}

// ParseJson parses bytes into array of struct.
func ParseJson(body []byte) (map[string]bool, error) {
	sites := make([]Site, 0)
	whiteList := make(map[string]bool)
	if err := json.Unmarshal(body, &sites); err != nil {
		log.Printf("Cannot parse site IP json files.")
		return whiteList, errors.New("Cannot parse site IP json files.")
	}

	for _, site := range sites {
		if site.Ipv4 != "" {
			whiteList[site.Ipv4] = true
		}
		if site.Ipv6 != "" {
			whiteList[site.Ipv6] = true
		}
	}
	return whiteList, nil
}

// LoadWhitelist load the site IP json from GCS.
func (ec *EmbargoCheck) LoadSiteIPJson() error {
	project := os.Getenv("GCLOUD_PROJECT")
	log.Printf("Using project: %s\n", project)
	json_url := SITE_IP_URL_TEST
	if project == "mlab-oti" {
		json_url = SITE_IP_URL
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

	ec.Whitelist, err = ParseJson(body)
	return err
}

// ReadWhitelistFromLocal loads IP whitelist from a local file.
func (ec *EmbargoCheck) ReadWhitelistFromLocal(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	whiteList := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		oneLine := strings.TrimSuffix(scanner.Text(), "\n")
		whiteList[oneLine] = true
	}
	ec.Whitelist = whiteList
	return nil
}

// CheckInWhitelist checks whether the IP in fileName is in the whitelist.
// It always returns true for non-web100 files.
// The filename is like: 20170225T23:00:00Z_4.34.58.34_0.web100
// file with IP that is in the IP whitelist, return true
// file with IP not in the whitelist, return false
func (ec *EmbargoCheck) CheckInWhitelist(fileName string) bool {
	if !strings.Contains(fileName, "web100") {
		return true
	}

	fn := FileName{Name: fileName}
	localIP := fn.GetLocalIP()
	return ec.Whitelist[localIP]
}
