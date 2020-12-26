/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dgrpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
)

func GetClientIP(ctx context.Context) string {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	if pr.Addr == net.Addr(nil) {
		return ""
	}

	addSlice := strings.Split(pr.Addr.String(), ":")
	if len(addSlice) < 1 {
		return ""
	}
	if addSlice[0] == "[" {
		return "127.0.0.1"
	}
	return addSlice[0]
}

func ShouldFail4Code(c codes.Code) bool {
	return c == codes.ResourceExhausted ||
		c == codes.Unknown ||
		c == codes.Internal ||
		c == codes.Unimplemented ||
		c == codes.DeadlineExceeded
}
