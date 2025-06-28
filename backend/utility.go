package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

func reconstructPath(end *PathNode) []GraphNode {
	var path []GraphNode
	for node := end; node != nil; node = node.Parent {
		path = append([]GraphNode{graph[strconv.Itoa(node.ID)]}, path...)
	}
	return path
}

func getGraphPath(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Tried to get path")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req PathGraphRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	path := aStarGraph(req.StartID, req.EndID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(path)
}

func loadGraph(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open graph file: %v", err)
	}
	defer file.Close()
	graph = make(map[string]GraphNode)

	if err := json.NewDecoder(file).Decode(&graph); err != nil {
		log.Fatalf("Failed to load graph: %v", err)
	}
	log.Printf("Successfully loaded graph with %d nodes\n", len(graph))
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371e3 // Earth's radius in meters

	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c

}

func findNearestNode(lat, lon float64) string {
	var nearestID string
	minDist := math.MaxFloat64

	for id, node := range graph {
		if len(node.Neighbors) < 2 { // avoid dead ends
			continue
		}
		d := haversine(lat, lon, node.Lat, node.Lon)
		if d < minDist {
			minDist = d
			nearestID = id
		}
	}
	return nearestID
}

func getRandomNodeID() string {
	connected := []string{}
	for id, node := range graph {
		if len(node.Neighbors) > 0 {
			connected = append(connected, id)
		}
	}
	if len(connected) == 0 {
		return "" // no valid nodes
	}
	return connected[rand.Intn(len(connected))]
}

func aStarGraphCoords(startLat, startLon, endLat, endLon float64) []GraphNode {
	startID := findNearestNode(startLat, startLon)
	endID := findNearestNode(endLat, endLon)
	return aStarGraph(startID, endID)
}
