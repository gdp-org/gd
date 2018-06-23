/**
 * Created by JetBrains GoLand.
 * Author: Chuck Chen
 * Date: 2018/6/22
 * Time: 14:59
 */

package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"bytes"
	"io"
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

func ParseJSON(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
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
	return json.Unmarshal(data, v)
}