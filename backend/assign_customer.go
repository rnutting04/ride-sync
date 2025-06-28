package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AssignCustomerRequest struct {
	DriverName string   `json:"driverName"`
	Customer   Customer `json:"customer"`
}

func assignCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AssignCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	driverMutex.Lock()
	defer driverMutex.Unlock()

	for i := range driverList {
		if driverList[i].Name == req.DriverName {

			fmt.Println("Found a path")
			driverList[i].HasCustomer = true
			driverList[i].Customer = req.Customer
			driverList[i].OnPickupLeg = true
			driverList[i].DestLat = req.Customer.Lat
			driverList[i].DestLon = req.Customer.Lon
			path := aStarGraphCoords(driverList[i].Lat, driverList[i].Lon, req.Customer.Lat, req.Customer.Lon)
			driverList[i].GraphPath = path
			driverList[i].PathIndex = 0
			driver := driverList[i]
			driverList[i].ETA = estimateETA(driverList[i].GraphPath)
			json.NewEncoder(w).Encode(driver.GraphPath)
			fmt.Println(driver.ETA)
			for _, node := range path {
				key := fmt.Sprintf("%.5f,%.5f", node.Lat, node.Lon) // Round to reduce duplicates
				heatmapCounts[key]++
			}
			break
		}

	}

	// ✅ Remove this customer from the queue (if still there)
	queueMutex.Lock()
	for i, c := range customerQueue {
		if c.Id == req.Customer.Id {
			customerQueue = append(customerQueue[:i], customerQueue[i+1:]...)
			break
		}
	}
	queueMutex.Unlock()

}
