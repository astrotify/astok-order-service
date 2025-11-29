package grpc

import (
	"fmt"
	"log"
	"net"
	orderGrpc "order-service/go-proto/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGRPCServer(port int, handler *OrderGrpcHandler) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	orderGrpc.RegisterOrderGRPCServiceServer(s, handler)

	// Enable reflection for testing with grpcurl
	reflection.Register(s)

	log.Printf("🚀 gRPC server listening on :%d", port)

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
