package main

import (
	"log"
	"mcenter/app"
	"net"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement customer.CustomerServer.
type server struct {
	savedCustomers []*app.CustomerRequest
}

// CreateCustomer creates a new Customer
func (s *server) CreateCustomer(ctx context.Context, in *app.CustomerRequest) (*app.CustomerResponse, error) {
	s.savedCustomers = append(s.savedCustomers, in)
	return &app.CustomerResponse{Id: in.Id, Success: true}, nil
}

// GetCustomers returns all customers by given filter
func (s *server) GetCustomers(filter *app.CustomerFilter, stream app.Customer_GetCustomersServer) error {
	for _, customer := range s.savedCustomers {
		if filter.Keyword != "" {
			if !strings.Contains(customer.Name, filter.Keyword) {
				continue
			}
		}
		if err := stream.Send(customer); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Creates a new gRPC server
	s := grpc.NewServer()
	app.RegisterCustomerServer(s, &server{})
	s.Serve(lis)
}
