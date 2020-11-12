package loadavg

import (
	"encoding/json"
	"log"
	"net/http"
)

func getavg() (*Stats, error) {
	return get()
}

// Stats represents load average values
type Stats struct {
	HostName                      string
	Loadavg1, Loadavg5, Loadavg15 float64
}

// returns json for avarage load
func getloadavgJSON() []byte {
	load, _ := getavg()
	l, err := json.Marshal(load)
	if err != nil {
		log.Println(err)
	}
	return l
}

// GetAvgLoad returns avarage cpu load
func GetAvgLoad(w http.ResponseWriter, r *http.Request) {
	payload := getloadavgJSON()
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}
