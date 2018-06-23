/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package error

import (
	"fmt"
)

var (
	ERR_CODE_SUCCESS       = 200
	ERR_CODE_PARA_ERROR    = 600
	ERR_CODE_APP_NOT_FOUND = 603
	ERR_CODE_SYS_ERROR     = 500
	ERR_CODE_DB_ERROR      = 701

	ERR_STR_UNKNOWN_ERROR = "unknown error"

	ErrMap = map[int]string{
		ERR_CODE_SUCCESS:    "ok",
		ERR_CODE_PARA_ERROR: "para error",
		ERR_CODE_DB_ERROR:   "db error",
		ERR_CODE_SYS_ERROR:  "system error",
	}
)

func GetErrorType(code int) string {
	t, ok := ErrMap[code]
	if !ok {
		t = ERR_STR_UNKNOWN_ERROR
	}
	return t
}

type MError struct {
	errCode int
	errType string
	errMsg  string
}

func (err *MError) Code() int {
	return err.errCode
}

func (err *MError) Type() string {
	return err.errType
}

func (err *MError) Error() string {
	return err.errMsg
}

func (err *MError) Detail() string {
	if err.errCode == ERR_CODE_SUCCESS {
		return err.Error()
	} else {
		return fmt.Sprintf("Type: %s, Error: %s", err.Type(), err.Error())
	}
}

func (err *MError) String() string {
	if err.errCode == ERR_CODE_SUCCESS {
		return fmt.Sprintf("Code: %d, Type: %s, Info: %s", err.Code(), err.Type(), err.Error())
	} else {
		return fmt.Sprintf("Code: %d, Type: %s, Error: %s", err.Code(), err.Type(), err.Error())
	}
}

func (err *MError) ToString() []byte {
	return []byte(fmt.Sprintf("MaeError[%s]", err.Detail()))
}

func NewHttpError(code int, format string, args ...interface{}) *MError {
	eType := GetErrorType(code)
	msg := format
	if len(format) > 0 && len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	err := &MError{
		errCode: code,
		errType: eType,
		errMsg:  msg,
	}

	return err
}

func MakeHttpError(code int, e error) *MError {
	eType := GetErrorType(code)
	err := &MError{
		errCode: code,
		errType: eType,
		errMsg:  e.Error(),
	}

	return err
}
