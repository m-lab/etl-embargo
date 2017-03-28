// Parse filename and return componants like log-time, IP, etc.
// Filename example: 20170315T01:00:00Z_173.205.3.39_0.web100
package embargo

import (
	"strings"
)

type FileName struct {
  name string
}

// GetLocalIP parse the filename and return IP. For old format, it will return empty string.
func (f *FileName) GetLocalIP() string {
	localIPStart := strings.IndexByte(f.name, '_')
	localIPEnd := strings.LastIndexByte(f.name, '_')
	if localIPStart < 0 || localIPEnd < 0 || localIPStart >= localIPEnd {
		return ""
	}
	return f.name[localIPStart+1 : localIPEnd]
}

func (f *FileName) GetDate() string {
  return f.name[0 : 8]
}

type FileNameParser interface {
  GetLocalIP()
  GetDate()
}


