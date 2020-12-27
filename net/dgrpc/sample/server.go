/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"context"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/net/dgrpc"
	pb "github.com/chuck1024/gd/net/dgrpc/sample/helloworld"
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
	rc := &mockReg{
		handler: &server{},
	}
	s := &dgrpc.GrpcServer{
		GrpcRunPort:     10240,
		RegisterHandler: rc,
		ServiceName:     "gd",
		UseTls:          true,
	}
	err := s.Run()
	if err != nil {
		return
	}
	dlog.Debug("err:%v", err)
	<-i
}
