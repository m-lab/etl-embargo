// Implement whitelist loading and embargo check based on filename.
package embargo

import (
	"bufio"
	"errors"
	"log"
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

// ReadWhitelistFromGCS loads IP whitelist from cloud storage.
func (ec *EmbargoCheck) ReadWhitelistFromGCS(bucket string, path string) bool {
	// TODO: Create service in a Singleton object, and reuse them for all GCS requests.
	checkService := CreateService()
	if checkService == nil {
		log.Printf("Storage service was not initialized.\n")
		return false
	}
	whiteList := make(map[string]bool)
	if fileContent, err := checkService.Objects.Get(bucket, path).Download(); err == nil {
		scanner := bufio.NewScanner(fileContent.Body)
		for scanner.Scan() {
			oneLine := strings.TrimSuffix(scanner.Text(), "\n")
			whiteList[oneLine] = true
		}
		ec.Whitelist = whiteList
		return true
	}
	return false
}

// ReadWhitelistFromLocal loads IP whitelist from a local file.
func (ec *EmbargoCheck) ReadWhitelistFromLocal(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	whiteList := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		oneLine := strings.TrimSuffix(scanner.Text(), "\n")
		whiteList[oneLine] = true
	}
	ec.Whitelist = whiteList
	return true
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
