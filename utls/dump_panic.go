/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package utls

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

var (
	dumpFlag   = os.O_CREATE | os.O_WRONLY
	dumpMode   = os.FileMode(0777)
	dumpPrefix = "dumpPanic."
)

func ReviewDumpPanic(file *os.File) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fileInfo.Size() == 0 {
		file.Close()
		return os.Remove(file.Name())
	}
	return nil
}

func Dump(name string) (*os.File, error) {
	suffix := fmt.Sprintf("-dump-%s", name)
	filename := dumpPrefix + suffix + "." + strconv.Itoa(os.Getpid())
	file, err := os.OpenFile(filename, dumpFlag, dumpMode)
	if err != nil {
		return file, err
	}

	if err := syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		return file, err
	}
	return file, nil
}
