package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

func getPairing(w http.ResponseWriter, r *http.Request) {
	type Pairing struct {
		IdealDriver     int        `json:"idealDriver"`
		CurrentCustomer Customer   `json:"currentCustomer"`
		Drivers         []Driver   `json:"drivers"`
		CustQue         []Customer `json:"custque"`
	}

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

	var requestData PairingRequest

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Replace with your frontend domain
	w.WriteHeader(http.StatusOK)

	queueMutex.Lock()
	var customer Customer
	if len(customerQueue) > 0 {
		customer = customerQueue[0] // Peek instead of dequeue
	}
	queueMutex.Unlock()
	pairing := Pairing{
		IdealDriver:     -1,
		CurrentCustomer: customer,
		Drivers:         requestData.Drivers,
		CustQue:         customerQueue,
	}
	leastDistance := 50000
	for i := 0; i < len(requestData.Drivers); i++ {
		dis := math.Abs(float64(requestData.Drivers[i].Lat-customer.Lat)) + math.Abs(float64(requestData.Drivers[i].Lon-customer.Lon))

		if int(dis) < leastDistance && !requestData.Drivers[i].HasCustomer {

			leastDistance = int(dis)
			pairing.IdealDriver = i

		}
	}
	if pairing.IdealDriver != -1 {
		requestData.Drivers[pairing.IdealDriver].HasCustomer = true
	} else {
		fmt.Println("No Drivers Available you are in position ", len(customerQueue))
	}
	fmt.Println(customer)
	json.NewEncoder(w).Encode(pairing)
}
