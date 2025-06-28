package main

import "strconv"

func estimateETA(path []GraphNode) float64 {
	totalSeconds := 0.0
	for i := 1; i < len(path); i++ {
		from := path[i-1]
		to := path[i]

		edge, ok := graph[strconv.Itoa(from.ID)].Neighbors[strconv.Itoa(to.ID)]
		if !ok || edge.Speed == 0 {
			continue
		}

		distance := haversine(from.Lat, from.Lon, to.Lat, to.Lon)
		seconds := (distance / (edge.Speed * 1000)) * 3600
		totalSeconds += seconds

		// Add realistic delay estimates
		if to.TrafficLight { // 30% chance to stop
			totalSeconds += 25 * .3 // average red light delay in seconds
		} else if to.StopSign { // 60% chance to stop
			totalSeconds += 5 * .6
		}
	}
	return totalSeconds / 60.0 // convert to minutes
}
