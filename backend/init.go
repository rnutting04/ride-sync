package main

import "sync"

var driverList = []Driver{}

var customerQueue []Customer
var queueMutex sync.Mutex
var driversInitialized bool = false
var graph map[string]GraphNode
var driverMutex sync.Mutex
var heatmapCounts = map[string]int{}
