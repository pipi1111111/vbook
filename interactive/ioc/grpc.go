package ioc

import (
	"google.golang.org/grpc"
	grpc2 "vbook/interactive/grpc"
	"vbook/pkg/grpcx"
)

func NewGrpcXServer(interSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	s := grpc.NewServer()
	interSvc.Register(s)
	return &grpcx.Server{
		Server: s,
		Addr:   ":8090",
	}
}
