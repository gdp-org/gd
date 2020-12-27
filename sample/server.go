/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package main

import (
	"context"
	"github.com/chuck1024/gd"
	de "github.com/chuck1024/gd/derror"
	"github.com/chuck1024/gd/net/dhttp"
	"github.com/chuck1024/gd/net/dogrpc"
	pb "github.com/chuck1024/gd/sample/helloworld"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"net/http"
)

type TestReq struct {
	Data string
}

type TestResp struct {
	Ret string
}

func HandlerHttpTest(c *gin.Context, req *TestReq) (code int, message string, err error, ret *TestResp) {
	gd.Debug("httpServerTest req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return http.StatusOK, "ok", nil, ret
}

func HandlerRpcTest(req *TestReq) (code uint32, message string, err error, ret *TestResp) {
	gd.Debug("rpc sever req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return uint32(de.RpcSuccess), "ok", nil, ret
}

type reg struct {
	handler pb.GreeterServer
}

func (r *reg) RegisterHandler(s *grpc.Server) error {
	pb.RegisterGreeterServer(s, r.handler)
	return nil
}

// server is used to implement hello world.GreeterServer.
type server struct{}

// SayHello implements hello world.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func Register(e *gd.Engine) {
	// http
	e.SetHttpServer(func(g *gin.Engine) error {
		r := g.Group("")
		r.Use(
			dhttp.GlFilter(),
			dhttp.StatFilter(),
			dhttp.GroupFilter(),
			dhttp.Logger("sample"),
		)

		e.HttpServer.POST(r, "test", HandlerHttpTest)

		if err := e.HttpServer.CheckHandle(); err != nil {
			return err
		}

		return nil
	})

	// Rpc
	e.RpcServer.AddDogHandler(1024, HandlerRpcTest)
	if err := e.RpcServer.DogRpcRegister(); err != nil {
		gd.Error("DogRpcRegister occur error:%s", err)
		return
	}
	dogrpc.InitFilters([]dogrpc.Filter{&dogrpc.GlFilter{}, &dogrpc.LogFilter{}})

	// grpc
	rc := &reg{
		handler: &server{},
	}
	e.SetGrpcServer(rc)
}

func main() {
	d := gd.Default()

	Register(d)

	err := d.Run()
	if err != nil {
		gd.Error("Error occurs, error = %s", err.Error())
		return
	}
}

// you can use command to test http service.
// curl -X POST http://127.0.0.1:10240/test -H "Content-Type: application/json" --data '{"Data":"test"}'
