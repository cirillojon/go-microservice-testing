# Basic calculation service using Go

This service supports basic mathematical operations such as: addition, subtraction, multiplication, and division. 

The service also logs each operation performed to a MongoDB Atlas database.

## Tools Used

1. **Go:** 
2. **MongoDB:**
3. **Go-Kit:** 
4. **Docker:**

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

Example: `{"a":5,"b":3,"op":"+"}`

Once the request is processed and the operation is performed, it is logged to a MongoDB Atlas database.

In case of an error (like division by zero or invalid operation), the service returns an error message in the response.

## How to Run and Test

**How to Run:**

go run calculater.go (requires .env values to be set)

**Test via Invoke-WebRequest**

Invoke-WebRequest -Uri http://localhost:8080/calculate -Method POST -ContentType "application/json" -Body '{"a":5,"b":3,"op":"+"}'

**How to Test using test file:**

go test (use -v for verbose output)

**How to Build Docker Image:**
docker build -t cirillojon/calculation-service .

## How to Run Docker Image:
docker run -e MONGO_USER=username -e MONGO_PASSWORD=password -e MONGO_DB_NAME=database -p localPort:8080 cirillojon/calculation-service

## Note
Remember to use the corresponding escape character for special characters in password 
