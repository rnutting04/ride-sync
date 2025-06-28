package main

import (
	"log"
	"math"
	"math/rand"
	"strconv"
)

func aStarGraph(startID, endID string) []GraphNode {
	openSet := map[string]*PathNode{}
	closedSet := map[string]bool{}

	start := graph[startID]
	goal := graph[endID]

	startNode := &PathNode{ID: start.ID, G: 0, F: heuristic(start, goal)}
	openSet[strconv.Itoa(start.ID)] = startNode

	for len(openSet) > 0 {
		var current *PathNode
		for _, node := range openSet {
			if current == nil || node.F < current.F {
				current = node
			}
		}

		if strconv.Itoa(current.ID) == endID {
			return reconstructPath(current)
		}

		delete(openSet, strconv.Itoa(current.ID))
		closedSet[strconv.Itoa(current.ID)] = true

		for neighborIDStr, info := range graph[strconv.Itoa(current.ID)].Neighbors {
			if closedSet[neighborIDStr] {
				continue
			}

			// Default speed fallback if missing or 0
			speed := info.Speed
			if speed <= 0 {
				speed = 40.0 // Default speed in km/h
			}

			// Compute time cost (in hours)
			timeCost := info.Distance / (speed * 1000.0 / 3600.0) // convert to seconds

			delayPenalty := 0.0

			node := graph[neighborIDStr]
			if node.TrafficLight {
				// Only stop at ~1 in 3 lights (simulate catching green)
				if rand.Float64() < 0.33 {
					delayPenalty += 15.0 // seconds
				}
			} else if node.StopSign {
				delayPenalty += 1.0 // seconds, soft penalty for ETA only
			}

			tentativeG := current.G + timeCost + delayPenalty

			neighborID, err := strconv.Atoi(neighborIDStr)
			if err != nil {
				log.Printf("Invalid neighbor ID: %s", neighborIDStr)
				continue
			}

			neighbor, exists := openSet[neighborIDStr]
			if !exists || tentativeG < neighbor.G {
				newNode := &PathNode{
					ID:     neighborID,
					G:      tentativeG,
					F:      tentativeG + heuristic(graph[neighborIDStr], goal),
					Parent: current,
				}
				openSet[neighborIDStr] = newNode
			}
		}
	}
	return nil
}

func heuristic(a, b GraphNode) float64 {
	// Use Euclidean or Haversine distance
	dLat := a.Lat - b.Lat
	dLon := a.Lon - b.Lon
	return math.Sqrt(dLat*dLat + dLon*dLon)
}
