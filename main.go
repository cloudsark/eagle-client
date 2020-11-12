package main

import (
	"eagle-client/loadavg"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/cpu/load/avg", loadavg.GetAvgLoad).Methods("GET")

	http.Handle("/", router)

	log.Println("Listening on :3000")
	log.Fatal(http.ListenAndServe("0.0.0.0:3000", nil))

}
