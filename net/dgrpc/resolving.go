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
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServerConf struct {
	CaPemEncrypted     string
	ServerKeyEncrypted string
	ServerPemEncrypted string
}

func DefaultServer(serviceDesc grpc.ServiceDesc) (*grpc.Server, error) {
	return DefaultServerConf(serviceDesc, nil)
}

func DefaultServerConf(serviceDesc grpc.ServiceDesc, conf *ServerConf) (*grpc.Server, error) {
	ops := []InterceptorOption{
		WithGlInterceptor(),
		WithPerfCounterInterceptor(serviceDesc.ServiceName),
		interceptor.WithLogInterceptor(),
		interceptor.WithRecoveryInterceptor(nil),
	}

	options := GetOptionHolder(ops...)

	if conf != nil {
		if conf.CertServerName == "" {
			return nil, errors.New("need cert server name")
		}

		// load trusted ca
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			return nil, errors.New("failed to parse root certificate")
		}
		// load server pem
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("load server pem fail,err=%v", err)
		}
		tlsConfig := &tls.Config{
			RootCAs:   roots,
			ClientCAs: roots,
			Certificates: []tls.Certificate{
				cert,
			},
			ServerName: conf.CertServerName,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
		c := credentials.NewTLS(tlsConfig)
		s := grpc.NewServer(
			grpc.Creds(c),
			grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(options.StreamServerInterceptors...)),
			grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(options.UnaryServerInterceptors...)),
		)
		return s, nil
	}
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(options.StreamServerInterceptors...)),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(options.UnaryServerInterceptors...)),
	)
	return s, nil
}
