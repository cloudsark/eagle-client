package linux

import (
	"eagle-client/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/shirou/gopsutil/disk"
)

/*
collect disk I/O
cat /proc/diskstats sample:
   8       0 sda 620457 177403 25699376 382130 184711 889675 22258129 364294 0 307683 466109 0 0 0 0
   8       1 sda1 192 1 67074 342 7 1 64 4 0 124 268 0 0 0 0
   8       2 sda2 620222 177402 25628430 381763 161032 889674 22258065 248699 0 291558 362268 0 0 0 0
  11       0 sr0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
 253       0 dm-0 288306 0 21581006 300975 191249 0 15502108 341830 0 203691 642805 0 0 0 0
 253       1 dm-1 504653 0 4040880 203525 879676 0 7037408 3968071 0 136791 4171596 0 0 0 0

fields:
  1.  major number
  2.  minor number
  3.  device name
  4.  reads completed successfully
  5.  reads merged
  6.  sectors read
  7.  time spent reading(ms)
  8.  writes completed
  9.  writes merged
  10. sectors written
  11. time spent writting(ms)
  12. I/Os currently in progress
  13. time spent doing I/Os(ms)
  14. weighted time spent doint I/Os(ms)

  Kernel 4.18+ appends four more fields for discard
  15. discards completed successfully
  16. discards merged
  17.  sectors discarded
  18.  time spent discarding

  Kernel 5.5+ appends two more fields for flush requests:
  19. flush requests completed successfully
  20. time spent flushing

		For more details refer to https://www.kernel.org/doc/Documentation/admin-guide/iostats.rst
*/

func (p DiskUsageStruct) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

func (p DiskIOtruct) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

// DiskUsageStruct - volume usage data
// {'sda1': {'used': '28851', 'percent': 84.0, 'free': '5625', 'volume': '/dev/sda1', 'path': '/', 'total': '36236'}
type DiskUsageStruct struct {
	Hostname    string
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Fstype      string  `json:"fstype"`
	Total       string  `json:"total"`
	Free        string  `json:"free"`
	Used        string  `json:"used"`
	UsedPercent float64 `json:"percent"`
}

// DiskIOtruct - volume io data
type DiskIOtruct struct {
	Hostname   string
	Name       string `json:"name"`
	Path       string `json:"path"`
	Reads      uint64 `json:"reads"`
	Writes     uint64 `json:"writes"`
	ReadBytes  uint64 `json:"bytes.read"`
	WriteBytes uint64 `json:"bytes.write"`
	WriteTime  uint64 `json:"write_time"`
	ReadTime   uint64 `json:"read_time"`
}

// DiskUsageList - list of volume usage data
type DiskUsageList []DiskUsageStruct

// DiskIOList - list of volume io data
type DiskIOList []DiskIOtruct

var sdiskRE = regexp.MustCompile(`/dev/(sd[a-z])[0-9]?`)

// removableFs checks if the volume is removable
func removableFs(name string) bool {
	s := sdiskRE.FindStringSubmatch(name)
	if len(s) > 1 {
		b, err := ioutil.ReadFile("/sys/block/" + s[1] + "/removable")
		if err != nil {
			return false
		}
		return strings.Trim(string(b), "\n") == "1"
	}
	return false
}

// isPseudoFS checks if it is a valid volume
func isPseudoFS(name string) (res bool) {
	err := utils.ReadLine("/proc/filesystems", func(s string) error {
		ss := strings.Split(s, "\t")
		if len(ss) == 2 && ss[1] == name && ss[0] == "nodev" {
			res = true
		}
		return nil
	})
	if err != nil {
		fmt.Printf("can not read '/proc/filesystems': %v", err)
	}
	return
}

// DiskUsage - return a list with disk usage structs
func diskUsage() (DiskUsageList, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		fmt.Printf("Error getting disk usage info: %v", err)
	}

	var usage DiskUsageList

	for _, p := range parts {
		if _, err := os.Stat(p.Mountpoint); err == nil {
			du, err := disk.Usage(p.Mountpoint)
			if err != nil {
				fmt.Printf("Error getting disk usage for Mount: %v", err)
			}

			if !isPseudoFS(du.Fstype) && !removableFs(du.Path) {

				TotalMB, _ := utils.ConvertBytesTo(du.Total, "mb", 0)
				FreeMB, _ := utils.ConvertBytesTo(du.Free, "mb", 0)
				UsedMB, _ := utils.ConvertBytesTo(du.Used, "mb", 0)

				UsedPercent := 0.0
				if TotalMB > 0 && UsedMB > 0 {
					UsedPercent = (float64(du.Used) / float64(du.Total)) * 100.0
					UsedPercent, _ = utils.FloatDecimalPoint(UsedPercent, 2)
					DeviceName := strings.Replace(p.Device, "/dev/", "", -1)

					TotalMBFormatted, _ := utils.FloatToString(TotalMB)
					FreeMBFormatted, _ := utils.FloatToString(FreeMB)
					UsedMBFormatted, _ := utils.FloatToString(UsedMB)

					d := DiskUsageStruct{
						Hostname:    utils.GetHostName(),
						Name:        DeviceName,
						Path:        du.Path,
						Fstype:      du.Fstype,
						Total:       TotalMBFormatted,
						Free:        FreeMBFormatted,
						Used:        UsedMBFormatted,
						UsedPercent: UsedPercent,
					}

					usage = append(usage, d)

				}

			}
		}
	}

	return usage, err
}

func getdiskJSON() []byte {
	load, _ := diskUsage()
	l, err := json.Marshal(load)
	if err != nil {
		log.Println(err)
	}
	return l
}

// GetDiskUsage returns disk usage
func GetDiskUsage(w http.ResponseWriter, r *http.Request) {
	payload := getdiskJSON()
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}
