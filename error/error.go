/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package error

import (
	"fmt"
)

var (
	TcpSuccess     = 0
	Success        = 200
	BadRequest     = 400
	Unauthorized   = 401
	Forbidden      = 403
	NotFound       = 404
	SystemError    = 500
	ParameterError = 600
	DBError        = 701
	CacheError     = 702

	UnknownError = "unknown error"

	ErrMap = map[int]string{
		TcpSuccess:     "ok",
		Success:        "ok",
		BadRequest:     "bad request",
		Unauthorized:   "Unauthorized",
		Forbidden:      "Forbidden",
		NotFound:       "not found",
		SystemError:    "system error",
		ParameterError: "Parameter error",
		DBError:        "db error",
		CacheError:     "cache error",
	}
)

func GetErrorType(code int) string {
	t, ok := ErrMap[code]
	if !ok {
		t = UnknownError
	}
	return t
}

type CodeError struct {
	errCode int
	errType string
	errMsg  string
}

func (err *CodeError) Code() int {
	return err.errCode
}

func (err *CodeError) Type() string {
	return err.errType
}

func (err *CodeError) Error() string {
	return err.errMsg
}

func (err *CodeError) Detail() string {
	if err.errCode == Success || err.errCode == TcpSuccess {
		return err.Error()
	} else {
		return fmt.Sprintf("Type: %s, Error: %s", err.Type(), err.Error())
	}
}

func (err *CodeError) String() string {
	if err.errCode == Success || err.errCode == TcpSuccess {
		return fmt.Sprintf("Code: %d, Type: %s, Info: %s", err.Code(), err.Type(), err.Error())
	} else {
		return fmt.Sprintf("Code: %d, Type: %s, Error: %s", err.Code(), err.Type(), err.Error())
	}
}

func (err *CodeError) ToString() []byte {
	return []byte(fmt.Sprintf("MaeError[%s]", err.Detail()))
}

func (err *CodeError) SetMsg(msg string) *CodeError {
	err.errMsg = msg
	return err
}

func SetCodeType(code int, errType string) *CodeError {
	err := &CodeError{
		errCode: code,
		errType: errType,
	}
	return err
}

func NewCodeError(code int, format string, args ...interface{}) *CodeError {
	eType := GetErrorType(code)
	msg := format
	if len(format) > 0 && len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	err := &CodeError{
		errCode: code,
		errType: eType,
		errMsg:  msg,
	}

	return err
}

func MakeCodeError(code int, e error) *CodeError {
	eType := GetErrorType(code)
	err := &CodeError{
		errCode: code,
		errType: eType,
		errMsg:  e.Error(),
	}

	return err
}
