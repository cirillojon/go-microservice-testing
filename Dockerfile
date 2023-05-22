# Start from a base image that has Go installed
FROM docker.io/library/golang:1.20 AS builder

# Set the current working directory in the container
WORKDIR /app

# Copy the entire source code from the current directory to the container
COPY . .

# Build the Go application
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Second stage build for smaller image size
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary file from the builder stage
COPY --from=builder /app/main .

# Expose port 8080
EXPOSE 8080

# Run the application when the container starts
CMD ["./main"]
