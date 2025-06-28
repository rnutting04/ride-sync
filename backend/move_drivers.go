package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func moveDrivers() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C
		driverMutex.Lock()

		now := time.Now()

		for i := range driverList {
			driver := &driverList[i]

			// Skip if not yet time to move
			if now.Before(driver.MoveTime) {
				continue
			}

			// 🚗 Has a path and more steps
			if len(driver.GraphPath) > 0 && driver.PathIndex < len(driver.GraphPath) {
				// Current and next step
				var prev GraphNode
				if driver.PathIndex > 0 {
					prev = driver.GraphPath[driver.PathIndex-1]
				} else {
					prev = GraphNode{Lat: driver.Lat, Lon: driver.Lon}
				}
				next := driver.GraphPath[driver.PathIndex]

				// Move driver
				driver.Lat = next.Lat
				driver.Lon = next.Lon
				driver.PathIndex++

				// Compute distance, speed, and delay
				distance := haversine(prev.Lat, prev.Lon, next.Lat, next.Lon)

				edgeInfo := graph[strconv.Itoa(prev.ID)].Neighbors[strconv.Itoa(next.ID)]
				variation := 0.9 + rand.Float64()*0.2
				driver.CurrentSpeed = edgeInfo.Speed * variation

				if driver.CurrentSpeed <= 0 {

					driver.CurrentSpeed = 40.0 // fallback default
				}
				driver.ResourceLeft -= distance * 0.001 // Fuel usage (0.001 L per meter)
				if driver.ResourceLeft <= 0 {
					driver.ResourceLeft = 40.0
				}

				seconds := (distance / (driver.CurrentSpeed * 1000)) * 3600

				moveDelay := time.Duration(seconds * float64(time.Second))
				driver.AnimationTime = now.Add(moveDelay)
				// Apply pause AFTER animation at current node
				if next.TrafficLight && rand.Float64() < 0.3 {
					moveDelay += 25 * time.Second
				} else if next.StopSign && rand.Float64() < 0.7 {
					moveDelay += 5 * time.Second
				}

				driver.MoveTime = now.Add(moveDelay)

				// fmt.Println(driver.MoveTime)
				continue // Go to next driver
			}

			// ✅ Reached end of path — handle logic
			if driver.PathIndex >= len(driver.GraphPath) {
				if driver.HasCustomer && driver.OnPickupLeg {
					// Begin drop-off
					dest := driver.Customer
					path := aStarGraphCoords(driver.Lat, driver.Lon, dest.DestinationLat, dest.DestinationLon)
					driver.GraphPath = path
					driver.PathIndex = 0
					driver.OnPickupLeg = false
					for _, node := range path {
						key := fmt.Sprintf("%.5f,%.5f", node.Lat, node.Lon) // Round to reduce duplicates
						heatmapCounts[key]++
					}
					fmt.Printf("%s picked up %s — heading to drop-off\n", driver.Name, dest.Name)

				} else if driver.HasCustomer && !driver.OnPickupLeg {
					// Drop-off complete
					fmt.Printf("%s dropped off %s\n", driver.Name, driver.Customer.Name)
					driver.HasCustomer = false
					driver.Customer = Customer{}
					driver.OnPickupLeg = false

					// Resume roaming
					startID := findNearestNode(driver.Lat, driver.Lon)
					destID := getRandomNodeID()
					driver.GraphPath = aStarGraph(startID, destID)
					driver.PathIndex = 0

				} else {
					// Idle roaming
					startID := findNearestNode(driver.Lat, driver.Lon)
					destID := getRandomNodeID()
					driver.GraphPath = aStarGraph(startID, destID)
					for i := 0; i < 5 && len(driver.GraphPath) == 0; i++ {
						destID = getRandomNodeID()
						driver.GraphPath = aStarGraph(startID, destID)
					}
					if len(driver.GraphPath) == 0 {
						fmt.Printf("❌ Still no path for driver %s. Marking as idle.\n", driver.Name)
						driver.MoveTime = now.Add(2 * time.Second) // Retry later
						continue
					}

					driver.PathIndex = 0
				}

				// Schedule next move attempt after short delay
				driver.MoveTime = now.Add(2 * time.Second)
				driver.ETA = estimateETA(driver.GraphPath)
			}
		}

		driverMutex.Unlock()
	}
}
