package main

//  Invoke-WebRequest -Uri http://localhost:8080/calculate -Method POST -ContentType "application/json" -Body '{"a":5,"b":3,"op":"+"}'

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CalculationService struct {
	// Database client
	client *mongo.Client
}

type OperationLog struct {
	A     int
	B     int
	Op    string
	Value int
	Time  time.Time
}

type calculationRequest struct {
	A  int    `json:"a"`
	B  int    `json:"b"`
	Op string `json:"op"`
}

type calculationResponse struct {
	V   int    `json:"v"`
	Err string `json:"err,omitempty"`
}

// Create a method to log the operation to MongoDB
func (s CalculationService) LogOperation(ctx context.Context, op OperationLog) error {
	// Select the database and collection
	collection := s.client.Database("your-database-name").Collection("your-collection-name")

	// Insert the operation log
	_, err := collection.InsertOne(ctx, op)
	if err != nil {
		log.Printf("Could not log operation: %v", err)
		return err
	}

	return nil
}

func (CalculationService) Calculate(ctx context.Context, a int, b int, op string) (int, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, errors.New("Cannot divide by zero")
		}
		return a / b, nil
	default:
		return 0, errors.New("Invalid operation")
	}
}

func decodeCalculationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request calculationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func makeCalculationEndpoint(s CalculationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(calculationRequest)
		v, err := s.Calculate(ctx, req.A, req.B, req.Op)
		if err != nil {
			return calculationResponse{V: v, Err: err.Error()}, nil
		}

		opLog := OperationLog{
			A:     req.A,
			B:     req.B,
			Op:    req.Op,
			Value: v,
			Time:  time.Now(),
		}
		// Log the operation
		if err := s.LogOperation(ctx, opLog); err != nil {
			log.Printf("Failed to log operation: %v", err)
		}

		return calculationResponse{V: v}, nil
	}
}

func main() {
	var httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	flag.Parse()

	// Load environment variables and connect to MongoDB
	client, err := connectMongoDB()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new CalculationService with the MongoDB client
	svc := CalculationService{
		client: client, // <- Here's the change
	}

	calculationHandler := httptransport.NewServer(
		makeCalculationEndpoint(svc),
		decodeCalculationRequest,
		encodeResponse,
	)

	http.Handle("/calculate", calculationHandler)
	errs := make(chan error)

	go func() {
		fmt.Println("Listening on port", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, nil)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println("Exit", <-errs)
}

func connectMongoDB() (*mongo.Client, error) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file: %w", err)
	}

	mongoUser := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDbName := os.Getenv("MONGO_DB_NAME")

	// Set client options
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb+srv://%s:%s@cluster0.vebhmxj.mongodb.net/%s?retryWrites=true&w=majority", mongoUser, mongoPassword, mongoDbName))

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to MongoDB: %w", err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to ping MongoDB: %w", err)
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}
