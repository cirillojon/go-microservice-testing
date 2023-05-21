package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httptransport "github.com/go-kit/kit/transport/http"
)

func TestCalculateHandler(t *testing.T) {
	svc := CalculationService{}
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
	}
}
