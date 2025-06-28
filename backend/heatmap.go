package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func handleHeatmapData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var heatPoints [][]interface{}
	for key, count := range heatmapCounts {
		parts := strings.Split(key, ",")
		lat, _ := strconv.ParseFloat(parts[0], 64)
		lon, _ := strconv.ParseFloat(parts[1], 64)
		heatPoints = append(heatPoints, []interface{}{lat, lon, count})
	}
	json.NewEncoder(w).Encode(heatPoints)
}
