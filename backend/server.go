package main

import (
	"fmt"
	"net/http"
)

func main() {
	loadGraph("graph/graph.json")
	fs := http.FileServer(http.Dir("frontend/"))
	http.Handle("/", fs)
	http.HandleFunc("/set-grid", setGrid)
	http.HandleFunc("/get-heatmap-data", handleHeatmapData)
	http.HandleFunc("/get-customer", getCustomer)
	http.HandleFunc("/get-pairing", getPairing)
	http.HandleFunc("/get-cust-que", getCustQ)
	http.HandleFunc("/get-drivers", getDrivers)
	http.HandleFunc("/assign-customer", assignCustomer)
	http.HandleFunc("/get-graph-path", getGraphPath)

	fmt.Println("Server running at :8080")

	http.ListenAndServe(":8080", nil)
}
