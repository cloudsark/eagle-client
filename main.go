package main

import (
	"eagle-client/metrics/linux"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var clientUsername = os.Getenv("CLIENT_USERNAME")
var clientPassword = os.Getenv("CLIENT_PASSWORD")

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/cpu/load/avg", use(linux.GetAvgLoad, basicAuth)).Methods("GET")
	router.HandleFunc("/api/v1/disk/usage/stat", use(linux.GetDiskUsage, basicAuth)).Methods("GET")

	http.Handle("/", router)

	log.Println("Listening on :10052")
	log.Fatal(http.ListenAndServe("0.0.0.0:10052", nil))

}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		if pair[0] != clientUsername || pair[1] != clientPassword {
			http.Error(w, "Not authorized", 401)
			return
		}

		h.ServeHTTP(w, r)
	}
}
