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
	"github.com/chuck1024/gd/config"
	log "github.com/chuck1024/gd/dlog"
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
	GrpcRunPort     int
	RegisterHandler IRegisterHandler

	UseTls             bool
	GrpcCertServerName string // if not, default gd
	GrpcCaPemFile      string
	GrpcServerKeyFile  string
	GrpcServerPemFile  string
}

func (s *GrpcServer) Run() error {
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

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.GrpcRunPort))
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
			log.Crash("start server fail,addr=%v,err=%v", s.GrpcRunPort, err)
		}
	}()

	return nil
}

func (s *GrpcServer) Stop() {
	s.closeOnce.Do(func() {
		s.s.GracefulStop()
	})
}

func (s *GrpcServer) DefaultServer() (*grpc.Server, error) {
	ops := []InterceptorOption{
		WithGlInterceptor(),
		WithPerfCounterInterceptor(config.Config().Section("Server").Key("serverName").String()),
		WithLogInterceptor(),
		WithRecoveryInterceptor(nil),
	}

	options := GetOptionHolder(ops...)

	if s.UseTls {
		if s.GrpcCaPemFile == "" {
			s.GrpcCaPemFile = "ca_pem.json"
		}

		if s.GrpcServerKeyFile == "" {
			s.GrpcServerKeyFile = "server_key.json"
		}

		if s.GrpcServerPemFile == "" {
			s.GrpcServerPemFile = "server_pem.json"
		}

		if s.GrpcCertServerName == "" {
			s.GrpcCertServerName = DefaultCertServiceName
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
		ServerName:   s.GrpcCertServerName,
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
