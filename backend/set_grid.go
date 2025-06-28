package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

func setGrid(w http.ResponseWriter, r *http.Request) {
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

	if !driversInitialized {
		driverList = []Driver{}
		names := []string{"Foe", "Joe", "Poe", "Doe", "Bow", "Crow", "Low", "Bro", "Flow", "Row", "Glo", "Oh"}
		nodeKeys := make([]string, 0, len(graph))
		for k := range graph {
			nodeKeys = append(nodeKeys, k)
		}

		if len(nodeKeys) == 0 {
			log.Println("❌ Cannot assign start/end keys: graph is empty.")
			http.Error(w, "Graph data is empty. Cannot assign drivers.", http.StatusInternalServerError)
			return
		}

		for _, name := range names {
			startKey := nodeKeys[rand.Intn(len(nodeKeys))]
			endKey := nodeKeys[rand.Intn(len(nodeKeys))]

			start := graph[startKey]
			end := graph[endKey]
			path := aStarGraphCoords(start.Lat, start.Lon, end.Lat, end.Lon)
			driver := Driver{
				Name:         name,
				Lat:          start.Lat,
				Lon:          start.Lon,
				DestLat:      end.Lat,
				DestLon:      end.Lon,
				Dir:          "down",
				OnPickupLeg:  false,
				GraphPath:    path,
				PathIndex:    0,
				ResourceLeft: 40.0,
				CurrentSpeed: 30.0,
				ETA:          estimateETA(path),
			}
			driverList = append(driverList, driver)

		}
		driversInitialized = true
		fmt.Println("Drivers initialized")
		fmt.Println("moveDrivers started, driver count:", len(driverList))
		go moveDrivers()
	} else {
		fmt.Println("Drivers already initialized — skipping re-init")
	}

	w.WriteHeader(http.StatusOK)
	fmt.Println("Sim Initialized")
}
