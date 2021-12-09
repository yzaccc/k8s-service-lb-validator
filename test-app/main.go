package main

import (
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", probe)
	http.ListenAndServe(":3000", nil)
}

func probe(w http.ResponseWriter, r *http.Request) {
	hostName, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("pod-name", hostName)
	w.WriteHeader(200)
}