package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

func enqueue(customer Customer) {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	customerQueue = append(customerQueue, customer)
}

func getCustomer(w http.ResponseWriter, r *http.Request) {

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

	// var requestData CustomerRequest

	// err := json.NewDecoder(r.Body).Decode(&requestData)
	// if err != nil {
	// 	http.Error(w, "Invalid request data", http.StatusBadRequest)
	// 	return
	// }

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Replace with your frontend domain
	w.WriteHeader(http.StatusOK)
	names := []string{"Ryan", "Luke", "Nancy", "Bob", "Jess"}
	customer := Customer{
		Id:             rand.Intn(1000) + 1,
		Name:           names[rand.Intn(5)],
		Lon:            0,
		Lat:            0,
		DestinationLon: 0,
		DestinationLat: 0,
	}

	custLocationID := getRandomNodeID()
	custDestID := getRandomNodeID()
	path := aStarGraph(custLocationID, custDestID)
	if len(path) == 0 {
		fmt.Printf("⚠️ Customer %s could not find path to random start/end node\n", customer.Name)
		// Try a new random destination, up to N retries
		for i := 0; i < 5; i++ {
			custLocationID = getRandomNodeID()
			custDestID = getRandomNodeID()
			path = aStarGraph(custLocationID, custDestID)
			if len(path) > 0 {
				break
			}
		}
	}
	if len(path) == 0 {
		// Still nothing — flag driver as idle and avoid updating state
		fmt.Printf("⚠️ Customer %s could not find path to random start/end node after five tries\n", customer.Name)

		return
	}

	custLocationNode := graph[custLocationID]
	customer.Lon = custLocationNode.Lon
	customer.Lat = custLocationNode.Lat
	fmt.Println(customer, "first")

	custDestNode := graph[custDestID]
	customer.DestinationLon = custDestNode.Lon
	customer.DestinationLat = custDestNode.Lat
	enqueue(customer)
	fmt.Println(customerQueue)
	fmt.Println(customer)
	custreturn := CustStuff{
		Customer: customer,
		CustQue:  customerQueue,
	}
	json.NewEncoder(w).Encode(custreturn)
}
