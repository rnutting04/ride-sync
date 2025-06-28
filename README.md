# RideSync

**RideSync** is a real-time ride-sharing simulation platform. It features driver-customer interaction, pathfinding, ETA estimation, fuel consumption tracking, and an interactive map with an optional heatmap overlay for route frequency analysis.

## Live Demo

The latest deployed version is available at:  
[https://ride-sync.onrender.com](https://ride-sync.onrender.com)

---

## Features

- Real-time simulation of drivers and customers
- A* pathfinding using a graph of real-world road data
- Dynamic ETA and fuel tracking
- Toggleable heatmap to visualize route density
- Leaflet-based map interface
- Responsive UI built with Tailwind CSS
- Optional HTMX support for progressive enhancement

---

## Tech Stack

- **Go (Golang)** – backend simulation and API server
- **Leaflet.js** – map rendering and tile layers
- **Tailwind CSS** – utility-first responsive design
- **HTMX** – optional frontend enhancement (for interaction without full page reloads)
- **Docker** – containerized development and deployment

---

## Setup Instructions

### Prerequisites

- Docker and Docker Compose installed

### Run Locally

```bash
# Clone the repository
git clone https://github.com/your-username/ridesync.git
cd ridesync

# Start the application
docker-compose up --build
```
Visit http://localhost:8080 in your browser.

## Heatmap Integration
The frontend includes a toggle to show or hide a heatmap overlay, which is dynamically generated based on frequently traversed paths (e.g., driver routes to pickup and dropoff points).

This helps visualize route popularity and potential traffic hotspots within the simulated environment.

## Future Enhancements

- Integration with MongoDB or another persistent store for historical route tracking

- WebSocket support for real-time updates

- Expanded analytics dashboard
