/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	log "github.com/gdp-org/gd/dlog"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"sync"
)

const (
	DefaultCertServiceName = "gd"
)

func init() {
	SetGrpcLogger()
}

type IRegisterHandler interface {
	RegisterHandler(s *grpc.Server) error
}

type GrpcServer struct {
	s               *grpc.Server
	l               net.Listener
	startOnce       sync.Once
	closeOnce       sync.Once
	GrpcRunHost     int              `inject:"grpcRunHost"`
	RegisterHandler IRegisterHandler `inject:"registerHandler"`
	ServiceName     string           `inject:"serviceName"`

	UseTls            bool   `inject:"grpcUseTls" canNil:"true"`
	GrpcCaPemFile     string `inject:"grpcCaPemFile" canNil:"true"`
	GrpcServerKeyFile string `inject:"grpcServerKeyFile" canNil:"true"`
	GrpcServerPemFile string `inject:"grpcServerPemFile" canNil:"true"`
}

func (s *GrpcServer) Start() error {
	var err error
	s.startOnce.Do(func() {
		err = s.start()
	})
	return err
}

func (s *GrpcServer) start() error {
	server, err := s.DefaultServer()
	if err != nil {
		return fmt.Errorf("init server fail,err=%v", err)
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.GrpcRunHost))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.s = server
	s.l = l
	err = s.RegisterHandler.RegisterHandler(server)
	if err != nil {
		return err
	}

	reflection.Register(server)
	err = s.startRun()
	return err
}

func (s *GrpcServer) startRun() error {
	go func() {
		err := s.s.Serve(s.l)
		if err != nil {
			log.Crash("start server fail,addr=%v,err=%v", s.GrpcRunHost, err)
		}
	}()

	return nil
}

func (s *GrpcServer) Close() {
	s.closeOnce.Do(func() {
		s.s.GracefulStop()
	})
}

func (s *GrpcServer) DefaultServer() (*grpc.Server, error) {
	ops := []InterceptorOption{
		WithGlInterceptor(),
		WithPerfCounterInterceptor(s.ServiceName),
		WithLogInterceptor(),
		WithRecoveryInterceptor(nil),
	}

	options := GetOptionHolder(ops...)

	if s.UseTls {
		if s.GrpcCaPemFile == "" {
			s.GrpcCaPemFile = "conf/ca.pem"
		}

		if s.GrpcServerKeyFile == "" {
			s.GrpcServerKeyFile = "conf/server.key"
		}

		if s.GrpcServerPemFile == "" {
			s.GrpcServerPemFile = "conf/server.pem"
		}

		if s.ServiceName == "" {
			s.ServiceName = DefaultCertServiceName
		}

		c, err := s.GetCredentialsByCA()
		if err != nil {
			return nil, err
		}

		server := grpc.NewServer(
			grpc.Creds(c),
			grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(options.StreamServerInterceptors...)),
			grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(options.UnaryServerInterceptors...)),
		)
		return server, nil
	}

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(options.StreamServerInterceptors...)),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(options.UnaryServerInterceptors...)),
	)
	return server, nil
}

func (s *GrpcServer) GetCredentialsByCA() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(s.GrpcServerPemFile, s.GrpcServerKeyFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(s.GrpcCaPemFile)
	if err != nil {
		return nil, err
	}

	if !certPool.AppendCertsFromPEM(ca) {
		return nil, errors.New("certPool.AppendCertsFromPEM err")
	}

	c := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		RootCAs:      certPool,
		ServerName:   s.ServiceName,
	})

	return c, err
}

func (s *GrpcServer) GetTLSCredentials() (credentials.TransportCredentials, error) {
	c, err := credentials.NewServerTLSFromFile(s.GrpcServerPemFile, s.GrpcServerKeyFile)
	if err != nil {
		return nil, err
	}

	return c, err
}
