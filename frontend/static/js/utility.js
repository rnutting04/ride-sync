
export async function getCustomer(){
  return fetch('/get-customer', {
      method:"POST",
      headers:{'Content-Type': 'application/json'}
    })
    .then(response=>response.json())
    .then(data=> {
      console.log(data)
      return data
    })
}

export async function getCustomerQ(grid){
  const requestData = {
      grid:grid,
    }

  return fetch('/get-cust-que', {
      method:"POST",
      headers:{'Content-Type': 'application/json'},
      body:JSON.stringify(requestData)
    })
    .then(response=>response.json())
    .then(data=> {
      console.log(data)
      return data
    })
}

export async function getPairing(drivers){

  console.log(drivers)
  const requestData = {
    drivers:drivers,
  }

  return fetch('/get-pairing', {
      method:"POST",
      headers:{'Content-Type': 'application/json'},
      body:JSON.stringify(requestData)
    })
    .then(response=>response.json())
    .then(data=> {
      console.log(data)
      return data
    })
}