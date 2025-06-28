package main

import (
	"encoding/json"
	"net/http"
)

func getCustQ(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		// Handle preflight requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS") // Add allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type CustQ struct {
		CustQ []Customer `json:"custque"`
	}
	var requestData CordRequest

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Replace with your frontend domain
	w.WriteHeader(http.StatusOK)
	custq := CustQ{
		CustQ: customerQueue,
	}

	json.NewEncoder(w).Encode(custq)
}
