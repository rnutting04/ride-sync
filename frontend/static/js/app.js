


import {getCustomer} from "./utility.js";
import {getPairing} from "./utility.js";


async function main(){

  document.addEventListener("DOMContentLoaded", function () {
    const bell = document.getElementById("notification-bell");
    const list = document.getElementById("notification-log");

    if (bell && list) {
      bell.addEventListener("click", () => {

        list.classList.toggle("hidden"); // Assumes you are using a "hidden" class to show/hide
      });

      // Optional: Close the list when clicking outside
      // document.addEventListener("click", (e) => {
      //   if (!bell.contains(e.target) && !list.contains(e.target)) {
      //     list.classList.add("hidden");
      //   }
      // });
    }
  });
  const map = L.map('map').setView([37.7616, -122.4232], 13); // Coordinates for Tampa, FL
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
  attribution:
    '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
}).addTo(map);
  map.zoomControl.setPosition('bottomright');


  fetch("/set-grid", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({}), // empty body or dummy payload
  }).then(res => {
    if (!res.ok) {
      console.error("Failed to initialize grid/drivers");
    } else {
      console.log("Grid/drivers initialized");
    }
  });
         
    
  const custRes = await fetch("/get-cust-que", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({})
  });

  const custData = await custRes.json();
  const custque = custData.custque || [];
    


  function createCustomerElements(lat, lon, destLat, destLon, custid) {
    const custIcon = L.icon({
      iconUrl: 'static/assets/customer.png',
      iconSize: [38, 38],
    });

    const destIcon = L.icon({
      iconUrl: 'static/assets/destination.png',
      iconSize: [24, 24],
    });

    const pickupMarker = L.marker([lat, lon], { icon: custIcon }).addTo(map);
    const destMarker = L.marker([destLat, destLon], { icon: destIcon }).addTo(map);

    customerMarkers[custid] = {
      pickup: pickupMarker,
      dest: destMarker
    };
  }

  function showNotification(message) {
  const box = document.getElementById("floating-notification");
  box.textContent = message;
  box.style.opacity = 1;

  // Fade out after 3 seconds
  setTimeout(() => {
    box.style.opacity = 0;
  }, 3000);

  // Also add to log
  const log = document.getElementById("notification-list");
  const li = document.createElement("li");
  li.textContent = `${new Date().toLocaleTimeString()} - ${message}`;
  log.prepend(li);
}


function renderCustomerPanel(customers) {
  const container = document.getElementById('customer-list');
  container.innerHTML = '';

  customers.forEach(cust => {
    const card = document.createElement('div');
    card.className = 'customer-card';

    card.innerHTML = `
      <div class="avatar"></div>
      <div class="info">
        <div class="name">${cust.name}</div>
        <div class="details">
          Pickup: (${cust.lat.toFixed(4)}, ${cust.lon.toFixed(4)})<br>
          Destination: (${cust.destinationLat.toFixed(4)}, ${cust.destinationLon.toFixed(4)})
        </div>
      </div>
    `;

    container.appendChild(card);
  });
}


function renderDriverPanel(drivers) {
  const container = document.getElementById('driver-panel');
  container.innerHTML = '';

  // Sort by status priority
  const statusPriority = { idle: 0, enroute: 1, busy: 2, offline: 3 };
  drivers.sort((a, b) => statusPriority[a.status] - statusPriority[b.status]);

  drivers.forEach(driver => {


    // Set status and task
    if (!driver.hasCustomer) {
      driver.status = "idle";
      driver.task = "Available";
    } else if (driver.onPickupLeg) {
      driver.status = "enroute";
      driver.task = `Picking up ${driver.customer.name}`;
    } else {
      driver.status = "busy";
      driver.task = `Dropping off ${driver.customer.name}`;
    }

    const card = document.createElement('div');
    card.className = 'driver-card';
    card.onclick = () => map.setView([driver.lat, driver.lon], 16);

    const speedText = driver.currentSpeed ? `${driver.currentSpeed.toFixed(1)} km/h` : '—';
    const fuelText = driver.resourceLeft !== undefined ? `${driver.resourceLeft.toFixed(1)} L` : '—';

    card.innerHTML = `
      <div class="avatar"></div>
      <div class="info">
        <div class="name">${driver.name}</div>
        <div class="details">
          ${driver.task || '—'}<br>
          ${driver.eta ? `ETA: ${Math.round(driver.eta)} min` : ''}<br>
          Fuel: ${fuelText} Speed: ${speedText}
        </div>
      </div>
      <div class="driver-status ${driver.status}">${driver.status}</div>
    `;
    container.appendChild(card);
  });
}

    
    
    let currentCustomer = {};
 
    
  // 1. SET GRID FIRST
  const lastLatLngMap = {};



  const drawnQueueIds = new Set(); // Track added queue entries

  const driverMarkers = {};          // name → L.marker
  const driverPolylineMap = {};      // name → L.polyline
  const customerMarkers = {};        // id → { pickup, dest }
  const lastCustomers = {};          // name → customer id
  const assignedCustomerIds = new Set();
  const driverStates = {}; // { [driverName]: { phase: "idle" | "enroute" | "pickup" | "dropoff" } }
  for(let i=0;i< custque.length;i++){
    createCustomerElements(custque[i].lat, custque[i].lon, 
      custque[i].destinationLat,custque[i].destinationLon, custque[i].id)
  }
  setInterval(async () => {

    const res = await fetch("/get-drivers");
    const drivers = await res.json();
    const activeQueueIds = new Set();
    const queueList = document.getElementById("queue-list");
    
    renderDriverPanel(drivers)
    if (!queueList) return;
    assignedCustomerIds.clear();
   
    drivers.forEach(driver => {

      // Marker setup
      if (driver.hasCustomer && driver.customer && driver.customer.id) {
            const custId = driver.customer.id;
            assignedCustomerIds.add(custId);
            // ...rest of your code
          }
      if (!driverMarkers[driver.name]) {

      const marker = L.marker([driver.lat, driver.lon], {
                icon: L.divIcon({
                  className: 'driver-div-icon',
                  html: `<img id="car-${driver.name}" src="static/assets/car.png" style="width: 38px; height: 20px; transform: rotate(0deg); transform-origin: center center;" />`,
                  iconSize: [38, 20],
                  iconAnchor: [19, 10]
                })
              }).addTo(map);

          driverMarkers[driver.name] = marker;
      }
      const marker = driverMarkers[driver.name];
      if (marker) {
        
          const oldLatLng = marker.getLatLng();
          const newLatLng = [driver.lat, driver.lon];
          const dummy = { lat: oldLatLng.lat, lng: oldLatLng.lng };
          const moveDuration = Math.max((new Date(driver.animationTime) - Date.now()) / 1000, 0.2);
  
          gsap.to(dummy, {
            lat: newLatLng[0],
            lng: newLatLng[1],
            duration: moveDuration,
            ease: "linear",
            onUpdate: () => {
              marker.setLatLng([dummy.lat, dummy.lng]);
            }
          });
          const dx = newLatLng[1] - oldLatLng.lng;
          const dy = newLatLng[0] - oldLatLng.lat;

          const distanceMoved = Math.sqrt(dx * dx + dy * dy);

          // Only rotate if moved enough (tune the threshold — try 0.00005 or 0.0001)
          const moveCheck = 0.00001
      
          if (distanceMoved > moveCheck) {
            const rawAngle = Math.atan2(dy, dx) * (180 / Math.PI) + 180;
            const correctedAngle = -rawAngle;

            const iconEl = document.getElementById(`car-${driver.name}`);
            if (iconEl) {
              const currentRotation = parseFloat(iconEl.dataset.angle || "0");
              let delta = correctedAngle - currentRotation;

              // Normalize rotation direction
              if (delta > 180) delta -= 360;
              if (delta < -180) delta += 360;

              const newRotation = currentRotation + delta;

              gsap.to(iconEl, {
                rotation: newRotation,
                duration: 0.25,
                ease: "linear",
                onUpdate: function () {
                  iconEl.style.transform = `rotate(${this.targets()[0].rotation}deg)`;
                }
              });

              iconEl.dataset.angle = newRotation;
            }
          }

          lastLatLngMap[driver.name] = {
            lat: driver.lat,
            lng: driver.lon
          };
          // Only draw if path is valid
        if (driver.graphPath && driver.graphPath.length > 1) {
          const latLngs = driver.graphPath.map(p => [p.lat, p.lon]);

          // If path already exists, update it
          if (driverPolylineMap[driver.name]) {
            driverPolylineMap[driver.name].setLatLngs(latLngs);
          } else {
            // Otherwise, draw it for the first time
            const polyline = L.polyline(latLngs, { color: "black" }).addTo(map);
            driverPolylineMap[driver.name] = polyline;
          }
        }

        // If driver no longer has a path, remove it
        if ((!driver.graphPath || driver.graphPath.length <= 1) && driverPolylineMap[driver.name]) {
          map.removeLayer(driverPolylineMap[driver.name]);
          delete driverPolylineMap[driver.name];
        }
      }
      if (!driverStates[driver.name]) {
        driverStates[driver.name] = { phase: "idle" };
      }

      const currentState = driverStates[driver.name];

      if (driver.hasCustomer && driver.onPickupLeg && currentState.phase !== "enroute") {
        showNotification(`🛣️ ${driver.name} en route to pick up ${driver.customer.name}`);
        currentState.phase = "enroute";
      }

      if (driver.hasCustomer && !driver.onPickupLeg && currentState.phase !== "pickup") {
        showNotification(`🧍 ${driver.name} picked up ${driver.customer.name}`);
        currentState.phase = "pickup";
      }

      if (!driver.hasCustomer && currentState.phase === "pickup") {
        showNotification(`🛬 ${driver.name} dropped off ${driver.customer?.name || "customer"}`);
        currentState.phase = "idle";
      }


      // Update marker location
      // driverMarkers[driver.name].setLatLng([driver.lat, driver.lon]);

      // Path draw
      if (driver.hasCustomer && driver.customer && driver.customer.id) {
        const custId = driver.customer.id;
        lastCustomers[driver.name] = custId;
        assignedCustomerIds.add(custId);
        activeQueueIds.add(`queue-${custId}`);

        // Create pickup/dest markers
        if (!customerMarkers[custId]) {
          createCustomerElements(driver.customer.lat, driver.customer.lon, driver.customer.destinationLat, driver.customer.destinationLon, custId)
        }
        if (!lastCustomers[driver.name] && driver.hasCustomer && driver.customer) {
         
        }
                if (
          driver.hasCustomer &&
          driver.customer &&
          lastCustomers[driver.name] &&
          !driver.onPickupLeg // transition to drop-off
        ) {
  
        }
        // Status text
      }

      // Clean up after drop-off
      if (!driver.hasCustomer && lastCustomers[driver.name]) {
        const custId = lastCustomers[driver.name];
        if (customerMarkers[custId]) {
          map.removeLayer(customerMarkers[custId].pickup);
          map.removeLayer(customerMarkers[custId].dest);
          delete customerMarkers[custId];
        }
        if (driverPolylineMap[driver.name]) {
          map.removeLayer(driverPolylineMap[driver.name]);
          delete driverPolylineMap[driver.name];
          driver.pathDrawn = false;
        }
        document.getElementById(`queue-${custId}`)?.remove();
        delete lastCustomers[driver.name];
      }
    });

}, 1000);


setInterval(async () => {
  const custRes = await fetch("/get-cust-que", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({})
  });

  const custData = await custRes.json();
  const custque = custData.custque || [];
  const queueList = document.getElementById("queue-list");
  if (!queueList) return;
  renderCustomerPanel(custque)
}, 200);

  document.getElementById("acceptcustomer").addEventListener("click", handleNodeAccept);
    async function handleNodeAccept(){
      let idealDriver;
      const res = await fetch("/get-drivers");
      const drivers = await res.json();
      let pairing = await getPairing(drivers)
      currentCustomer = pairing.currentCustomer
     
      idealDriver = drivers[pairing.idealDriver]
      console.log(currentCustomer)
      await fetch("/assign-customer", {
        method: "POST",
        body: JSON.stringify({customer: currentCustomer,  driverName: idealDriver.name }),
        headers: { "Content-Type": "application/json" }
      }).then(response=>response.json())
      .then(data=> {
        console.log(data)
        return data
      })
  

      
   }
  document.getElementById("customerping").addEventListener("click", handleNodeSpawn);
  async function handleNodeSpawn(){

      let queCustomer = await getCustomer()
  
      console.log(queCustomer.custque)
      
  
      for(let i=0;i<queCustomer.custque.length;i++){
        if (!customerMarkers[queCustomer.custque[i].id]) {
          createCustomerElements(queCustomer.custque[i].lat, queCustomer.custque[i].lon, 
            queCustomer.custque[i].destinationLat,queCustomer.custque[i].destinationLon, queCustomer.custque[i].id)
        }


      } 
      
    }

    const sidePanel = document.getElementById("side-panel");
    const hamburger = document.getElementById("hamburger");
    console.log(hamburger)
    hamburger.addEventListener("click", () => {
      document.body.classList.toggle("side-panel-visible");
      sidePanel.classList.toggle("hidden");

    });



  }
      
       

main();
  

