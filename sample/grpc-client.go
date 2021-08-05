/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"context"
	"fmt"
	"github.com/gdp-org/gd"
	pb "github.com/gdp-org/gd/sample/helloworld"
	"google.golang.org/grpc"
	"strconv"
	"time"
)

// server is used to implement hello world.GreeterServer.
type server2 struct{}

// SayHello implements hello world.GreeterServer
func (s *server2) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	bc := gd.NewGrpcClient("127.0.0.1:10242", func(conn *grpc.ClientConn) (interface{}, error) {
		rawClient := pb.NewGreeterClient(conn)
		return rawClient, nil
	}, "gd")
	defer bc.Stop()

	c := bc.GetRawClient().(pb.GreeterClient)
	name := "test " + strconv.FormatInt(time.Now().Unix(), 10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	fmt.Println(fmt.Sprintf("Greeting: %s, err=%v", r, err))
}
