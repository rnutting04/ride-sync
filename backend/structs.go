package main

import "time"

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

type PathGraphRequest struct {
	StartID string `json:"startID"`
	EndID   string `json:"endID"`
}

type GraphNode struct {
	ID           int                     `json:"id"`
	Lat          float64                 `json:"lat"`
	Lon          float64                 `json:"lon"`
	Neighbors    map[string]NeighborInfo `json:"neighbors"` // keyed by neighbor node ID
	TrafficLight bool                    `json:"traffic_light"`
	StopSign     bool                    `json:"stop_sign"`
}
type PathRequest struct {
	Grid      [][]GridObject `json:"grid"`
	StartX    int            `json:"startX"`
	StartY    int            `json:"startY"`
	EndCordsX int            `json:"endCordsX"`
	EndCordsY int            `json:"endCordsY"`
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

type PathNode struct {
	ID     int
	G      float64
	F      float64
	Parent *PathNode
}

type NeighborInfo struct {
	Distance float64 `json:"distance"`
	Speed    float64 `json:"speed"` // km/h, optional
}
