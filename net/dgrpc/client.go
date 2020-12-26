/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"sync"
	"time"
)

type GrpcClient struct {
	Target             string
	Timeout            int
	MaxRetries         int
	PerCallTimeoutInMs int
	RetryCodes         []codes.Code // 指定重试的状态码，默认情况下只有codes.ResourceExhausted, codes.Unavailable才会重试
	WaitReady          bool
	UseTls             bool
	CertServerName     string
	GrpcCaPemFile      string
	GrpcClientKeyFile  string
	GrpcClientPemFile  string
	startOnce          sync.Once
	stopOnce           sync.Once
	connect            *grpc.ClientConn
	rawClient          interface{}
}

func (c *GrpcClient) Start(makeRawClient func(conn *grpc.ClientConn) (interface{}, error), getServiceDesc func() grpc.ServiceDesc) error {
	var err error
	c.startOnce.Do(func() {
		err = c.start(makeRawClient, getServiceDesc)
	})
	return err
}

func (c *GrpcClient) start(makeRawClient func(conn *grpc.ClientConn) (interface{}, error), getServiceDesc func() grpc.ServiceDesc) (err error) {
	c.connect, err = c.DefaultClient(getServiceDesc())
	if err != nil {
		return err
	}

	c.rawClient, err = makeRawClient(c.connect)
	if err != nil {
		return err
	}
	return nil
}

func (c *GrpcClient) Stop() {
	c.stopOnce.Do(func() {
		c.close()
	})
}

func (c *GrpcClient) close() {
	c.connect.Close()
}

func (c *GrpcClient) GetRawClient() interface{} {
	return c.rawClient
}

func (c *GrpcClient) DefaultClient(serviceDesc grpc.ServiceDesc) (*grpc.ClientConn, error) {
	ops := []InterceptorOption{
		WithGlInterceptor(),
		WithPerfCounterInterceptor(serviceDesc.ServiceName),
	}

	if c.Timeout > 0 {
		ops = append(ops, WithClientTimeOutInterceptor(time.Duration(c.Timeout)))
	} else {
		ops = append(ops, WithClientTimeOutInterceptor(1*time.Second))
	}

	if c.MaxRetries > 0 {
		if c.RetryCodes == nil || len(c.RetryCodes) == 0 {
			c.RetryCodes = grpcRetry.DefaultRetriableCodes
		}

		ops = append(ops, WithRetryInterceptor(
			grpcRetry.WithMax(uint(c.MaxRetries)),
			grpcRetry.WithPerRetryTimeout(time.Duration(c.PerCallTimeoutInMs)*time.Millisecond),
			grpcRetry.WithCodes(c.RetryCodes...),
		))
	}

	options := GetOptionHolder(ops...)

	to, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cc *grpc.ClientConn
	var err error
	if c.UseTls {
		if c.GrpcCaPemFile == "" {
			c.GrpcCaPemFile = "ca_pem.json"
		}

		if c.GrpcClientKeyFile == "" {
			c.GrpcClientKeyFile = "client_key.json"
		}

		if c.GrpcClientPemFile == "" {
			c.GrpcClientPemFile = "miot-grpc_client_pem.json"
		}

		if c.CertServerName == "" {
			c.CertServerName = DefaultCertServiceName
		}

		cTls, err := c.GetCredentialsByCA()
		if err != nil {
			return nil, err
		}

		cc, err = grpc.Dial(
			c.Target,
			grpc.WithTransportCredentials(cTls),
			grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
			grpc.WithUnaryInterceptor(grpcMiddleware.ChainUnaryClient(
				options.UnaryClientInterceptors...,
			)),
			grpc.WithStreamInterceptor(grpcMiddleware.ChainStreamClient(
				options.StreamClientInterceptors...,
			)),
		)
	} else {
		cc, err = grpc.Dial(
			c.Target,
			grpc.WithInsecure(),
			grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
			grpc.WithUnaryInterceptor(grpcMiddleware.ChainUnaryClient(
				options.UnaryClientInterceptors...,
			)),
			grpc.WithStreamInterceptor(grpcMiddleware.ChainStreamClient(
				options.StreamClientInterceptors...,
			)),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("grpc dail fail,target=%v,err=%v", serviceDesc.ServiceName, err)
	}

	if c.WaitReady {
		err := c.WaitClientReady(to, cc)
		if err != nil {
			return nil, fmt.Errorf("wait client ready fail,target=%v,err=%v", serviceDesc.ServiceName, err)
		}
	}
	return cc, nil
}

func (c *GrpcClient) WaitClientReady(ctx context.Context, cc *grpc.ClientConn) error {
	for {
		s := cc.GetState()
		if s == connectivity.Ready {
			break
		}
		if !cc.WaitForStateChange(ctx, s) {
			return ctx.Err()
		}
	}
	return nil
}

func (c *GrpcClient) GetCredentialsByCA() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(c.GrpcClientPemFile, c.GrpcClientKeyFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(c.CertServerName)
	if err != nil {
		return nil, err
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, err
	}

	t := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &cert, nil
		},
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return &cert, nil
		},
		ServerName: c.CertServerName,
		RootCAs:    certPool,
		ClientCAs:  certPool,
	})

	return t, err
}

func (c *GrpcClient) GetTLSCredentials() (credentials.TransportCredentials, error) {
	t, err := credentials.NewClientTLSFromFile(c.GrpcClientPemFile, c.GrpcClientKeyFile)
	if err != nil {
		return nil, err
	}

	return t, err
}
