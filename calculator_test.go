package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCalculateHandler(t *testing.T) {
	// Create a MongoDB client for testing
	client, err := createTestMongoClient()
	if err != nil {
		t.Fatal(err)
	}

	// Cleanup the test client at the end
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Fatal(err)
		}
	}()

	svc := CalculationService{
		client: client,
	}
	handler := httptransport.NewServer(makeCalculationEndpoint(svc), decodeCalculationRequest, encodeResponse)

	tests := []struct {
		a, b     int
		op       string
		expected calculationResponse
	}{
		{1, 2, "+", calculationResponse{V: 3}},
		{5, 3, "-", calculationResponse{V: 2}},
		// More test cases...
	}

	for _, tt := range tests {
		// Create an HTTP request with the input for this test case
		input := calculationRequest{A: tt.a, B: tt.b, Op: tt.op}
		requestBody, _ := json.Marshal(input)
		req := httptest.NewRequest("POST", "/calculate", bytes.NewBuffer(requestBody))

		// Use httptest to record the HTTP response
		w := httptest.NewRecorder()

		// Send the HTTP request to our handler
		handler.ServeHTTP(w, req)

		// Check that the HTTP response status is 200 OK
		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// Check that the HTTP response body is correct
		var response calculationResponse
		json.NewDecoder(w.Body).Decode(&response)
		if response != tt.expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				response, tt.expected)
		}

		// Check if the operation log is inserted into the MongoDB collection
		collection := client.Database("your-database-name").Collection("your-collection-name")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var log OperationLog
		err := collection.FindOne(ctx, OperationLog{A: tt.a, B: tt.b, Op: tt.op}).Decode(&log)
		if err != nil {
			t.Errorf("failed to find operation log in MongoDB: %v", err)
		}
	}
}

func createTestMongoClient() (*mongo.Client, error) {
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

	return client, nil
}
