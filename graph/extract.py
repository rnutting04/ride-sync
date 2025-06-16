import json

# Load the OSMnx-style graph
with open("sf_graph.json") as f:
    osm_data = json.load(f)

nodes = osm_data["nodes"]
links = osm_data["links"]

graph = {}

# Default speeds by road type (km/h)
DEFAULT_SPEED_BY_HIGHWAY = {
    "motorway": 110,
    "motorway_link": 80,
    "trunk": 100,
    "trunk_link": 70,
    "primary": 90,
    "primary_link": 70,
    "secondary": 70,
    "secondary_link": 60,
    "tertiary": 60,
    "residential": 40,
    "living_street": 20,
    "service": 30,
    "unclassified": 40,
    "road": 40,
}

# Step 1: Initialize nodes (include traffic info)
for node in nodes:
    node_id = str(node["id"])
    highway_type = node.get("highway", "")

    graph[node_id] = {
        "id": node["id"],
        "lat": node["y"],
        "lon": node["x"],
        "neighbors": {},
        "traffic_light": highway_type == "traffic_signals",
        "stop_sign": highway_type == "stop"
    }

# Step 2: Parse maxspeed with fallback logic
def parse_maxspeed(link):
    def extract_one(val):
        if isinstance(val, str):
            if "mph" in val.lower():
                try:
                    return float(val.lower().replace("mph", "").strip()) * 1.60934  # mph → km/h
                except:
                    return None
            else:
                try:
                    return float(val.strip())  # Assume already in km/h
                except:
                    return None
        elif isinstance(val, (int, float)):
            return float(val)
        return None

    # Try maxspeed first
    ms = link.get("maxspeed")
    if isinstance(ms, list):
        for val in ms:
            parsed = extract_one(val)
            if parsed:
                return parsed
    elif ms:
        return extract_one(ms)

    # Fallback to default by highway type
    hwy = link.get("highway")
    if hwy and isinstance(hwy, str):
        return DEFAULT_SPEED_BY_HIGHWAY.get(hwy, 40)

    return 40  # Final fallback

# Step 3: Add neighbors with distance + inferred speed
for link in links:
    source = str(link["source"])
    target = str(link["target"])
    distance = link.get("length", 1.0)
    speed = parse_maxspeed(link)

    if source in graph:
        graph[source]["neighbors"][target] = {
            "distance": distance,
            "speed": speed
        }

# ✅ Save to file
with open("graph.json", "w") as f:
    json.dump(graph, f, indent=2)
