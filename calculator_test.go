package main

// To use: go test -v

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// MockMongoRepository is a struct that implements the DatabaseRepository interface, used for testing
type MockMongoRepository struct{}

// InsertOne is a method of MockMongoRepository that is required by the DatabaseRepository interface. This mock implementation doesn't do anything.
func (m MockMongoRepository) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	return nil, nil
}

// TestCalculate is a function that tests the Calculate method of the CalculationService.
// It uses different test cases with different operations and checks if the output matches the expected results.
func TestCalculate(t *testing.T) {
	// Creates a CalculationService with the mock repository for testing
	calc := CalculationService{
		dbRepo: MockMongoRepository{},
	}

	// Cases to test, each with a name, inputs a and b, operation, expected result and error
	cases := []struct {
		name   string
		a      int
		b      int
		op     string
		expect int
		err    error
	}{
		// Each case tests a different operation or error condition
		{name: "Addition", a: 1, b: 1, op: "+", expect: 2, err: nil},
		{name: "Subtraction", a: 5, b: 3, op: "-", expect: 2, err: nil},
		{name: "Multiplication", a: 2, b: 2, op: "*", expect: 4, err: nil},
		{name: "Division", a: 4, b: 2, op: "/", expect: 2, err: nil},
		{name: "Division by zero", a: 4, b: 0, op: "/", expect: 0, err: errors.New("cannot divide by zero")},
		{name: "Invalid operation", a: 4, b: 2, op: "$", expect: 0, err: errors.New("invalid operation")},
	}

	for _, tt := range cases {
		// Runs each case as a subtest
		t.Run(tt.name, func(t *testing.T) {
			res, err := calc.Calculate(context.Background(), tt.a, tt.b, tt.op)
			// Checks if the result and error match the expected result and error
			if res != tt.expect || (err != nil && err.Error() != tt.err.Error()) {
				t.Errorf("Expected %v and error %v, but got %v and error %v", tt.expect, tt.err, res, err)
			}
		})
	}
}

// TestLogOperation is a function that tests the LogOperation method of the CalculationService.
// It checks if the operation is logged correctly without any errors.
func TestLogOperation(t *testing.T) {
	// Creates a CalculationService with the mock repository for testing
	calc := CalculationService{
		dbRepo: MockMongoRepository{},
	}

	// Defines an operation to log
	op := OperationLog{
		A:     5,
		B:     3,
		Op:    "+",
		Value: 8,
		Time:  time.Now(),
	}

	err := calc.LogOperation(context.Background(), op)

	// Checks if the operation was logged successfully
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		fmt.Println("LogOperation passed successfully")
	}
}
