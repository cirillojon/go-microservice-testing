package main

// For testing:
// Invoke-WebRequest -Uri http://localhost:8080/calculate -Method POST -ContentType "application/json" -Body '{"a":5,"b":3,"op":"+"}'

// Import necessary libraries
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
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// DatabaseRepository defines the interface for any database that will be used with the CalculationService
type DatabaseRepository interface {
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
}

// MongoDBRepository is an implementation of DatabaseRepository that uses MongoDB as its database
type MongoDBRepository struct {
	client     *mongo.Client
	db         string
	collection string
}

// InsertOne implements the InsertOne method of DatabaseRepository for MongoDB
func (r MongoDBRepository) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	// Accesses the database collection and inserts the document
	collection := r.client.Database(r.db).Collection(r.collection)
	return collection.InsertOne(ctx, document)
}

// CalculationService is a service that provides calculation operations and operation logging
type CalculationService struct {
	dbRepo DatabaseRepository
}

// LogOperation logs the calculation operation into the database
func (s CalculationService) LogOperation(ctx context.Context, op OperationLog) error {
	// Calls the InsertOne method of the DatabaseRepository
	_, err := s.dbRepo.InsertOne(ctx, op)
	if err != nil {
		log.Printf("Could not log operation: %v", err)
		return err
	}
	return nil
}

// OperationLog is the structure of the calculation operation to be logged
type OperationLog struct {
	A     int
	B     int
	Op    string
	Value int
	Time  time.Time
}

// CalculationRequest is the structure of the calculation request from the client
type calculationRequest struct {
	A  int    `json:"a"`
	B  int    `json:"b"`
	Op string `json:"op"`
}

// CalculationResponse is the structure of the calculation response to the client
type calculationResponse struct {
	V   int    `json:"v"`
	Err string `json:"err,omitempty"` // Error message will be empty if there is no error
}

// Calculate performs the calculation operation and returns the result or an error
func (CalculationService) Calculate(ctx context.Context, a int, b int, op string) (int, error) {
	// Depending on the operation, performs the appropriate calculation
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			// Returns an error if the operation is division by zero
			return 0, errors.New("cannot divide by zero")
		}
		return a / b, nil
	default:
		// Returns an error if the operation is not supported
		return 0, errors.New("invalid operation")
	}
}

// decodeCalculationRequest decodes the calculation request from the client
func decodeCalculationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request calculationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Returns an error if the request cannot be decoded
		return nil, err
	}
	return request, nil
}

// encodeResponse encodes the response to the client
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// makeCalculationEndpoint makes an endpoint for the calculation operation
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
		log.Printf("Operation log before calling LogOperation: %+v", opLog)

		// Logs the calculation operation
		if err := s.LogOperation(ctx, opLog); err != nil {
			log.Printf("Failed to log operation: %v", err)
		}

		return calculationResponse{V: v}, nil
	}
}

// The main function starts the service
func main() {
	// Parses the command line arguments
	var httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	flag.Parse()

	// Loads the environment variables and connects to MongoDB
	client, err := connectMongoDB()
	if err != nil {
		log.Fatal(err)
	}

	// Creates a MongoDBRepository
	dbRepo := MongoDBRepository{
		client:     client,
		db:         "Go-Service",
		collection: "Calculator",
	}

	// Creates a CalculationService with the MongoDBRepository
	svc := CalculationService{
		dbRepo: &dbRepo,
	}

	// Creates a new server for the calculation endpoint
	calculationHandler := httptransport.NewServer(
		makeCalculationEndpoint(svc),
		decodeCalculationRequest,
		encodeResponse,
	)

	// Adds the calculation endpoint to the HTTP handler
	http.Handle("/calculate", calculationHandler)
	errs := make(chan error)

	// Starts the HTTP server in a goroutine so that it doesn't block
	go func() {
		fmt.Println("Listening on port", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, nil)
	}()

	// Listens for signals for graceful shutdown
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println("Exit", <-errs)
}

// connectMongoDB connects to MongoDB and returns a client
func connectMongoDB() (*mongo.Client, error) {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	mongoUser := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDbName := os.Getenv("MONGO_DB_NAME")

	// Set client options
	clientOptions := options.Client().ApplyURI(
		fmt.Sprintf("mongodb+srv://%s:%s@cluster0.vebhmxj.mongodb.net/%s?retryWrites=true&w=majority", mongoUser, mongoPassword, mongoDbName),
	).SetReadPreference(readpref.Primary())

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}
