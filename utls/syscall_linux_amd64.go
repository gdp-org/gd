//go:build linux && amd64
// +build linux,amd64

/**
 * Copyright 2021 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package utls

import (
	"syscall"
)

func Dup2(from int, to int) error {
	if err := syscall.Dup2(from, to); err != nil {
		return err
	}
	return nil
}
