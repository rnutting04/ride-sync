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
	"sync"
	"time"
)

var customerQueue []Customer
var queueMutex sync.Mutex
var driversInitialized bool = false
var graph map[string]GraphNode

type NeighborInfo struct {
	Distance float64 `json:"distance"`
	Speed    float64 `json:"speed"` // km/h, optional
}

type GraphNode struct {
	ID           int                     `json:"id"`
	Lat          float64                 `json:"lat"`
	Lon          float64                 `json:"lon"`
	Neighbors    map[string]NeighborInfo `json:"neighbors"` // keyed by neighbor node ID
	TrafficLight bool                    `json:"traffic_light"`
	StopSign     bool                    `json:"stop_sign"`
}

func loadGraph(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open graph file: %v", err)
	}
	defer file.Close()

	var loaded map[string]GraphNode
	if err := json.NewDecoder(file).Decode(&loaded); err != nil {
		log.Fatalf("Failed to load graph: %v", err)
	}

	graph = loaded
}

type GridObject struct {
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Heuristic    int    `json:"heuristic"`
	G            int    `json:"g"`
	F            int    `json:"f"`
	IsHazard     bool   `json:"isHazard"`
	TrafficLevel int    `json:"trafficLevel"`
	Type         string `json:"type"` // <-- add this
}

type PathRequest struct {
	Grid      [][]GridObject `json:"grid"`
	StartX    int            `json:"startX"`
	StartY    int            `json:"startY"`
	EndCordsX int            `json:"endCordsX"`
	EndCordsY int            `json:"endCordsY"`
}

type Node struct {
	X            int
	Y            int
	Heuristic    int
	G            int
	F            int
	Parent       *Node
	Visited      bool
	IsHazard     bool
	TrafficLevel int
}

type Cords struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Customer struct {
	Id             int     `json:"id"`
	Name           string  `json:"name"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	DestinationLat float64 `json:"destinationLat"`
	DestinationLon float64 `json:"destinationLon"`
}

type CustStuff struct {
	Customer Customer   `json:"customer"`
	CustQue  []Customer `json:"custque"`
}

type LatLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Driver struct {
	Name          string      `json:"name"`
	Dir           string      `json:"dir"`
	Rotation      int         `json:"rotation"`
	Lat           float64     `json:"lat"`     // instead of X
	Lon           float64     `json:"lon"`     // instead of Y
	DestLat       float64     `json:"destLat"` // instead of Destinationx
	DestLon       float64     `json:"destLon"` // instead of Destinationy
	Req           int         `json:"req"`
	HasCustomer   bool        `json:"hasCustomer"`
	Customer      Customer    `json:"customer"`
	GraphPath     []GraphNode `json:"graphPath"` // instead of [][]int
	PathIndex     int         `json:"pathIndex"`
	OnPickupLeg   bool        `json:"onPickupLeg"`
	ResourceLeft  float64     `json:"resourceLeft"`
	CurrentSpeed  float64     `json:"currentSpeed"`
	MoveTime      time.Time   `json:"moveTime"`
	AnimationTime time.Time   `json:"animationTime"`
	ETA           float64     `json:"eta"`
}

type CustomerRequest struct {
	Grid [][]GridObject `json:"grid"`
}

type DriverRequest struct {
	Grid   [][]GridObject `json:"grid"`
	Driver Driver         `json:"driver"`
}

type CordRequest struct {
	Grid [][]GridObject `json:"grid"`
}

type PairingRequest struct {
	Drivers []Driver `json:"drivers"`
}

var driverList = []Driver{}

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

func enqueue(customer Customer) {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	customerQueue = append(customerQueue, customer)
}

func get_rand_cords(grid [][]GridObject) (int, int) {
	for {
		x := rand.Intn(len(grid))
		y := rand.Intn(len(grid[0]))
		tile := grid[x][y]

		// Only spawn on roads or intersections
		if tile.Type == "road" || tile.Type == "intersection" {
			return x, y
		}
	}
}

type PathNode struct {
	ID     int
	G      float64
	F      float64
	Parent *PathNode
}

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

func reconstructPath(end *PathNode) []GraphNode {
	var path []GraphNode
	for node := end; node != nil; node = node.Parent {
		path = append([]GraphNode{graph[strconv.Itoa(node.ID)]}, path...)
	}
	return path
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

func getCords(w http.ResponseWriter, r *http.Request) {

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

	var requestData CordRequest

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Replace with your frontend domain
	w.WriteHeader(http.StatusOK)
	cords := Cords{
		X: 0,
		Y: 0,
	}

	cords.X, cords.Y = get_rand_cords(requestData.Grid)

	json.NewEncoder(w).Encode(cords)
}

func getCustQ(w http.ResponseWriter, r *http.Request) {

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
	type CustQ struct {
		CustQ []Customer `json:"custque"`
	}
	var requestData CordRequest

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Replace with your frontend domain
	w.WriteHeader(http.StatusOK)
	custq := CustQ{
		CustQ: customerQueue,
	}

	json.NewEncoder(w).Encode(custq)
}

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
	fmt.Println(requestData.Drivers)
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
			driverList[i].GraphPath = aStarGraphCoords(driverList[i].Lat, driverList[i].Lon, req.Customer.Lat, req.Customer.Lon)
			driverList[i].PathIndex = 0

			driver := driverList[i]
			driverList[i].ETA = estimateETA(driverList[i].GraphPath)
			json.NewEncoder(w).Encode(driver.GraphPath)
			fmt.Println(driver.ETA)
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

var driverMutex sync.Mutex

func getDrivers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	driverMutex.Lock()
	defer driverMutex.Unlock()

	err := json.NewEncoder(w).Encode(driverList)
	if err != nil {
		log.Printf("❌ Failed to encode driver list: %v\n", err)
	}
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
					driver.GraphPath = aStarGraphCoords(driver.Lat, driver.Lon, dest.DestinationLat, dest.DestinationLon)
					driver.PathIndex = 0
					driver.OnPickupLeg = false

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
		names := []string{"Foe", "Joe","Poe", "Doe", "Bow", "Crow", "Low", "Bro", "Flow", "Row", "Glo", "Oh"}
		nodeKeys := make([]string, 0, len(graph))
		for k := range graph {
			nodeKeys = append(nodeKeys, k)
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

// func loadGraphData() error {
// 	file, err := os.Open("graph.json")
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	decoder := json.NewDecoder(file)
// 	return decoder.Decode(&graph)
// }

type PathGraphRequest struct {
	StartID string `json:"startID"`
	EndID   string `json:"endID"`
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

func main() {
	loadGraph("graph/graph.json")
	// err := loadGraphData()
	// if err != nil {
	// 	log.Fatal("Failed to load graph:", err)
	// }
	fs := http.FileServer(http.Dir("frontend/"))
	http.Handle("/", fs)
	http.HandleFunc("/set-grid", setGrid)
	http.HandleFunc("/get-customer", getCustomer)
	http.HandleFunc("/get-cords", getCords)
	http.HandleFunc("/get-pairing", getPairing)
	http.HandleFunc("/get-cust-que", getCustQ)
	http.HandleFunc("/get-drivers", getDrivers)
	http.HandleFunc("/assign-customer", assignCustomer)
	http.HandleFunc("/get-graph-path", getGraphPath)

	fmt.Println("Server running at :8080")

	http.ListenAndServe(":8080", nil)
}
