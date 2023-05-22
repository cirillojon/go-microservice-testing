# Basic calculation service using Go

This service supports basic mathematical operations such as: addition, subtraction, multiplication, and division. 

The service also logs each operation performed to a MongoDB Atlas database.

## Tools Used

1. **Go:** 
2. **MongoDB:**
3. **Go-Kit:** 

## Code Functionality

The service exposes a single HTTP POST endpoint at `/calculate` which accepts JSON requests with the structure:

`json
{
  "a": <number>,
  "b": <number>,
  "op": <operation>
}
`

Where `op` is one of the following strings: "+", "-", "*", "/".

Once the request is processed and the operation is performed, it is logged to a MongoDB Atlas database.

In case of an error (like division by zero or invalid operation), the service returns an error message in the response.

## How to Run and Test

**How to Run:**

go run calculater.go (requires .env values to be set)

**Test via Invoke-Request**

Invoke-WebRequest -Uri http://localhost:8080/calculate -Method POST -ContentType "application/json" -Body '{"a":5,"b":3,"op":"+"}'

**How to Test using test file:**

go test (use -v for verbose output)
