// Parse filename and return componants like log-time, IP, etc.
// Filename example: 20170315T01:00:00Z_173.205.3.39_0.web100
package embargo

import (
	"strings"
)

// GetLocalIP parse the filename and return IP. For old format, it will return empty string.
func GetLocalIP(fileName string) string {
	localIPStart := strings.IndexByte(fileName, '_')
	localIPEnd := strings.LastIndexByte(fileName, '_')
	if localIPStart < 0 || localIPEnd < 0 || localIPStart >= localIPEnd {
		return ""
	}
	return fileName[localIPStart+1 : localIPEnd]
}

func GetDate(fileName string) string {
  return fileName[0 : 8]
}
