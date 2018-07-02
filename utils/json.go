/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
)

func loadFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func LoadJsonToObject(filename string, t interface{}) error {
	buf, e := loadFile(filename)
	if buf == nil || e != nil {
		return e
	}

	if err := json.Unmarshal(buf, &t); err != nil {
		return err
	}

	return nil
}

func ParseJSON(path string, v interface{}) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	mode := info.Mode()
	if mode.IsDir() {
		return errors.New("Invalid config file.it is directory. ")
	}

	if !mode.IsRegular() {
		return errors.New("Invalid config file,it is not a regular file. ")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var lines [][]byte
	buf := bytes.NewBuffer(data)
	for {
		line, err := buf.ReadBytes('\n')
		line = bytes.Trim(line, " \t\r\n")
		if len(line) > 0 && !bytes.HasPrefix(line, []byte("//")) {
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	data = bytes.Join(lines, []byte{})
	if err = json.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}
