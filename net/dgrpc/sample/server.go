/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"context"
	"github.com/gdp-org/gd"
	"github.com/gdp-org/gd/net/dgrpc"
	pb "github.com/gdp-org/gd/net/dgrpc/sample/helloworld"
	"google.golang.org/grpc"
)

type mockReg struct {
	handler pb.GreeterServer
}

func (r *mockReg) RegisterHandler(s *grpc.Server) error {
	pb.RegisterGreeterServer(s, r.handler)
	return nil
}

// server is used to implement hello world.GreeterServer.
type server struct{}

// SayHello implements hello world.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	var i chan struct{}
	defer gd.LogClose()
	rc := &mockReg{
		handler: &server{},
	}
	s := &dgrpc.GrpcServer{
		GrpcRunHost:     10240,
		RegisterHandler: rc,
		ServiceName:     "gd",
		UseTls:          true,
	}
	err := s.Start()
	if err != nil {
		return
	}
	gd.Debug("err:%v", err)
	<-i
}
