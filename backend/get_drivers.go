package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func getDrivers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	driverMutex.Lock()
	defer driverMutex.Unlock()

	err := json.NewEncoder(w).Encode(driverList)
	if err != nil {
		log.Printf("❌ Failed to encode driver list: %v\n", err)
	}
}
