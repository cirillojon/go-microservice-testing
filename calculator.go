package main

//  Invoke-WebRequest -Uri http://localhost:8080/calculate -Method POST -ContentType "application/json" -Body '{"a":5,"b":3,"op":"+"}'

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

type CalculationService struct{}

type calculationRequest struct {
	A  int    `json:"a"`
	B  int    `json:"b"`
	Op string `json:"op"`
}

type calculationResponse struct {
	V   int    `json:"v"`
	Err string `json:"err,omitempty"`
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
		return calculationResponse{V: v}, nil
	}
}

func main() {
	var httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	flag.Parse()

	svc := CalculationService{}

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
