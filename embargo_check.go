// Implement whitelist loading and embargo check based on filename.
package embargo

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type EmbargoCheck struct {
  whitelist map[string]bool
  embargodate string
}

// TODO: Read IP whitelist from Data Store.

// ReadWhitelistFromLocal load IP whitelist from a local file.
func (ec *EmbargoCheck) ReadWhitelistFromLocal(path string){
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	whiteList := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		oneLine := strings.TrimSuffix(scanner.Text(), "\n")
		whiteList[oneLine] = true
	}
	ec.whitelist = whiteList
}

// ReadWhitelistFromGCS load IP whitelist from cloud storage.
func (ec *EmbargoCheck) ReadWhitelistFromGCS(path string) {
        checkService := CreateService()
	if checkService == nil {
		fmt.Printf("Storage service was not initialized.\n")
		return
	}
	whiteList := make(map[string]bool)
	if fileContent, err := checkService.Objects.Get("sidestream-embargo", path).Download(); err == nil {
		scanner := bufio.NewScanner(fileContent.Body)
		for scanner.Scan() {
			oneLine := strings.TrimSuffix(scanner.Text(), "\n")
			whiteList[oneLine] = true
		}
		ec.whitelist = whiteList
	}
}

// EmbargoCheck decide whether to embargo it based on embargo date and IP
// whitelist given a filename of sidestream test.
// The filename is like: 20170225T23:00:00Z_4.34.58.34_0.web100.gz
// THe embargo date is like 20160225
// file with date on or before the embargo date are always published. Return false
// file with IP that is in the IP whitelist are always published. Return false
// file with date after the embargo date and IP not in the whitelist will be embargoed. Return true
func (ec *EmbargoCheck) ShouldEmbargo(fileName string) bool {
	date, err := strconv.Atoi(fileName[0:8])
	if err != nil {
		fmt.Println(err)
		return true
	}
	embargoDateInt, err := strconv.Atoi(ec.embargodate)
	if err != nil {
		fmt.Println(err)
		return true
	}
	if date < embargoDateInt {
		return false
	}
        fn := FileName{name: fileName}
	localIP := fn.GetLocalIP()
	// For old filename, that do not contain IP, always embargo them.
	if ec.whitelist[localIP] {
		return false
	}
	return true
}

type ReadWLer interface {
  ReadWhitelistFromLocal(path string)
  ReadWhitelistFromGCS(path string)
  ShouldEmbargo(fileName string)
}
