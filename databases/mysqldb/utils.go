/**
 * Copyright 2019 mysqldb Author. All rights reserved.
 * Author: Chuck1024
 */

package mysqldb

import (
	"bytes"
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	easyerrors "github.com/go-errors/errors"
	"net"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func IsDbConnError(err error) bool {
	if err == nil {
		return false
	} else if IsTimeoutError(err) || err == driver.ErrBadConn || err == mysql.ErrInvalidConn {
		return true
	} else if err == context.DeadlineExceeded || err == context.Canceled {
		return true
	} else if strings.Contains(err.Error(), "connection refused") {
		return true
	} else {
		mse, ok := err.(*mysql.MySQLError)
		if ok && mse != nil {
			//1062 is duplicate entry error
			if mse.Number == 1062 {
				return false
			}

			return true
		}
	}

	return false
}

func GetFieldsName(v interface{}) (res string, errRet error) {
	fieldNamesArray, errRet := GetFieldsNameArray(v)
	if errRet != nil {
		return
	}
	return strings.Join(fieldNamesArray, ","), nil
}

func GetFieldsNameArray(v interface{}) (res []string, errRet error) {
	WithRecover(
		func() {
			res, errRet = _getFieldsNameArray(v)
		},
		func(panicErr interface{}) {
			mixErr := fmt.Errorf("GetFieldsNameArray paniced!panic=%v,origErr=%v", panicErr, errRet)
			errRet = mixErr
		},
	)

	return
}

func _getFieldsNameArray(v interface{}) (res []string, errRet error) {
	res = make([]string, 0)
	typ := reflect.TypeOf(v)
	if typ == nil {
		errRet = fmt.Errorf("input cannot be nil")
		return
	}
	// we need to get the struct value if it is a pointer
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		errRet = fmt.Errorf("not a struct type")
		return
	}
	nf := typ.NumField()
	fieldNamesArray := make([]string, 0, nf)
	for i := 0; i < nf; i++ {
		tf := typ.Field(i)
		mysqlFieldName := tf.Tag.Get("mysqlField")
		if mysqlFieldName == "" {
			pbtags := strings.Split(tf.Tag.Get("protobuf"), ",")
			for _, pbtag := range pbtags {
				nametag := strings.Split(pbtag, "=")
				if len(nametag) > 1 && nametag[0] == "name" && nametag[1] != "" {
					mysqlFieldName = nametag[1]
					break
				}
			}
		}
		if mysqlFieldName != "" {
			mysqlFieldName = "`" + mysqlFieldName + "`"
			fieldNamesArray = append(fieldNamesArray, mysqlFieldName)
		}
	}

	return fieldNamesArray, nil
}

func GetFields(v interface{}) (res []interface{}, errRet error) {
	defer func() {
		if r := recover(); r != nil {
			res = nil
			errRet = fmt.Errorf("err when getFields in reflect %v", r)
		}
	}()
	if v == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// we need to get the struct value if it is a pointer
	if rv.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("input need to be a pointer")
	}
	rv = rv.Elem()
	num := rv.NumField()
	rt := rv.Type()
	res = make([]interface{}, 0, num)
	for i := 0; i < num; i++ {
		field := rv.Field(i)
		tField := rt.Field(i)
		tag := tField.Tag
		if tag.Get("mysqlField") != "" || tag.Get("protobuf") != "" {
			if !field.CanAddr() {
				return nil, fmt.Errorf("%v cannot be address", tField.Name)
			}
			res = append(res, field.Addr().Interface())
		}
	}
	return
}

func buildWhereSql(data map[string]interface{}) (string, []interface{}) {
	placeholder := make([]string, 0, len(data))
	valholder := []interface{}{}
	for idx, v := range data {
		k := MysqlEscapeString(idx)
		switch val := v.(type) {
		case []interface{}:
			valholder = append(valholder, val...)
			str := fmt.Sprintf(" `%s` in %s ", k, makeSqlPlaceHolderForIn(val))
			placeholder = append(placeholder, str)
		case []int64:
			temvarholder := make([]interface{}, 0, len(val))
			for _, v := range val {
				temvarholder = append(temvarholder, v)
			}
			valholder = append(valholder, temvarholder...)
			str := fmt.Sprintf(" `%s` in %s ", k, makeSqlPlaceHolderForIn(temvarholder))
			placeholder = append(placeholder, str)
		case []string:
			temvarholder := make([]interface{}, 0, len(val))
			for _, v := range val {
				temvarholder = append(temvarholder, v)
			}
			valholder = append(valholder, temvarholder...)
			str := fmt.Sprintf(" `%s` in %s ", k, makeSqlPlaceHolderForIn(temvarholder))
			placeholder = append(placeholder, str)
		default:
			valholder = append(valholder, val)
			str := fmt.Sprintf(" `%s` = ? ", k)
			placeholder = append(placeholder, str)
		}
	}
	return strings.Join(placeholder, "AND"), valholder
}

func makeSqlPlaceHolderForIn(vals []interface{}) string {
	buffer := bytes.NewBufferString("(")
	for k := range vals {
		buffer.WriteString("?")
		if k != len(vals)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString(")")
	return buffer.String()
}

type SqlCondition struct {
	tableprefix string
	conds       []*condition
	order       []*orderby
	offset      int64
	limit       int64
}

type condition struct {
	key     string
	compare string
	value   interface{}
}

type orderby struct {
	key  string
	desc bool
}

func NewSqlCondition() *SqlCondition {
	return &SqlCondition{
		conds: []*condition{},
		order: []*orderby{},
		limit: 300,
	}
}

func (c *SqlCondition) WithTablePrefix(tableprefix string) *SqlCondition {
	c.tableprefix = MysqlEscapeString(tableprefix)
	return c
}

func (c *SqlCondition) WithCondition(key, compare string, value interface{}) *SqlCondition {
	key = strings.Trim(key, " ")
	if key != "" {
		compare = strings.Trim(compare, " ")
		switch compare {
		case "=", ">", ">=", "<", "<=", "<>":
		default:
			compare = "="
		}
		c.conds = append(c.conds, &condition{
			key:     MysqlEscapeString(key),
			compare: compare,
			value:   value,
		})
	}
	return c
}

func (c *SqlCondition) WithOrder(key string, isdesc bool) *SqlCondition {
	key = strings.Trim(key, " ")
	key = "`" + key + "`"
	if key != "" {
		c.order = append(c.order, &orderby{
			key:  MysqlEscapeString(key),
			desc: isdesc,
		})
	}
	return c
}

func (c *SqlCondition) WithLimit(limit int64) *SqlCondition {
	if limit >= 0 {
		c.limit = limit
	}
	return c
}

func (c *SqlCondition) WithOffset(offset int64) *SqlCondition {
	if offset > 0 {
		c.offset = offset
	}
	return c
}

/*
	Valid SqlCondition,use before buildsql
*/
func (c *SqlCondition) Valid(isget bool) error {
	if !isget {
		if len(c.conds) == 0 {
			return errors.New("cant alter all table record once")
		}
	} else {
		if c.limit == 0 && len(c.conds) == 0 {
			return errors.New("cant load all table record once")
		}
	}
	return nil
}

/*
	returns:
	@sqlstr string      contains all contents after(include) WHERE
    @vars   []interface	varsholder
*/
func (c *SqlCondition) BuildWhereSql() (string, []interface{}) {
	_, str, vars := c.BuildShardWhereSql("")
	return str, vars
}

/*
	Need to call WithTablePrefix() first
	returns:
	@tablename	string	tablename
	@sqlstr string      contains all contents after(include) WHERE
    @vars	[]interface	varsholder
*/
func (c *SqlCondition) BuildShardWhereSql(shardkey string) (string, string, []interface{}) {
	//build table name
	tablename := MysqlEscapeString(c.tableprefix + shardkey)
	//build conditon str
	valholder := []interface{}{}
	placeholder := make([]string, 0, len(c.conds))
	for _, v := range c.conds {
		k := v.key
		switch val := v.value.(type) {
		case []interface{}:
			valholder = append(valholder, val...)
			str := fmt.Sprintf(" `%s` IN %s", k, makeSqlPlaceHolderForIn(val))
			placeholder = append(placeholder, str)
		case []int64:
			temvarholder := make([]interface{}, 0, len(val))
			for _, v := range val {
				temvarholder = append(temvarholder, v)
			}
			valholder = append(valholder, temvarholder...)
			str := fmt.Sprintf(" `%s` IN %s", k, makeSqlPlaceHolderForIn(temvarholder))
			placeholder = append(placeholder, str)
		case []string:
			temvarholder := make([]interface{}, 0, len(val))
			for _, v := range val {
				temvarholder = append(temvarholder, v)
			}
			valholder = append(valholder, temvarholder...)
			str := fmt.Sprintf(" `%s` IN %s", k, makeSqlPlaceHolderForIn(temvarholder))
			placeholder = append(placeholder, str)
		default:
			valholder = append(valholder, val)
			if v.compare == "" {
				v.compare = "="
			}
			str := fmt.Sprintf(" `%s` %s ?", k, v.compare)
			placeholder = append(placeholder, str)
		}
	}
	condstr := strings.Join(placeholder, " AND")
	//build order by
	if len(c.order) != 0 {
		orderholder := make([]string, 0, len(c.order))
		condstr = condstr + " ORDER BY "
		for _, v := range c.order {
			buffer := bytes.NewBufferString(v.key)
			if v.desc {
				buffer.WriteString(" DESC")
			} else {
				buffer.WriteString(" ASC")
			}
			orderholder = append(orderholder, buffer.String())
		}
		condstr = condstr + strings.Join(orderholder, ",")
	}
	//build limit
	if c.limit > 0 {
		condstr = condstr + " LIMIT ?"
		if c.offset > 0 {
			condstr = condstr + ",?"
			valholder = append(valholder, interface{}(c.offset))
		}
		valholder = append(valholder, interface{}(c.limit))
	}
	if condstr != "" && !strings.HasPrefix(condstr, " LIMIT ?") {
		condstr = " WHERE" + condstr
	}
	return tablename, condstr, valholder
}

func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case *url.Error:
		if err, ok := err.Err.(net.Error); ok && err.Timeout() {
			return true
		}
	case net.Error:
		if err.Timeout() {
			return true
		}
	default:
	}

	return false
}

func WithRecover(fn func(), errHandler func(interface{})) (err interface{}) {
	defer func() {
		if err = recover(); err != nil {
			wraped := easyerrors.Wrap(err, 2)
			stacktrace := wraped.ErrorStack()
			//string that starts with "panic_recovered: " in stderr will trigger
			//xbox sms GO_CORE alert
			fmt.Fprintln(os.Stderr, "panic_recovered:", stacktrace)
			if errHandler != nil {
				errHandler(err)
			}
		}
	}()

	fn()
	return
}

func MysqlEscapeString(source string) string {
	var j int = 0
	if len(source) == 0 {
		return ""
	}
	tempStr := source[:]
	desc := make([]byte, len(tempStr)*2)
	for i := 0; i < len(tempStr); i++ {
		var escape byte
		escape = 0
		switch tempStr[i] {
		case 0:
			escape = '0'
		case '\r':
			escape = 'r'
		case '\n':
			escape = 'n'
		case '\\':
			escape = '\\'
		case '\'':
			escape = '\''
		case '"':
			escape = '"'
		case '\032':
			escape = 'Z'
		}
		if escape != 0 {
			desc[j] = '\\'
			desc[j+1] = escape
			j = j + 2
		} else {
			desc[j] = tempStr[i]
			j = j + 1
		}
	}
	return string(desc[0:j])
}

// GetDatastructFields, 获取 datastructs 包中的 fields，并以 slice 形式返回
// data 参数为要获取 field 的实例指针
func GetDataStructFields(data interface{}) ([]string, error) {
	typeOf := reflect.TypeOf(data).Elem()
	numField := typeOf.NumField()
	fieldSlice := make([]string, 0, numField)
	for i := 0; i < numField; i++ {
		tfield := typeOf.Field(i)
		if len(tfield.PkgPath) > 0 {
			return nil, fmt.Errorf("field %s is not public", tfield.Name)
		}
		mysqlFieldName := tfield.Tag.Get("mysqlField")
		if len(mysqlFieldName) == 0 {
			return nil, fmt.Errorf("field %s has no mysqlField tag", tfield.Name)
		}
		fieldSlice = append(fieldSlice, "`"+mysqlFieldName+"`")
	}
	return fieldSlice, nil
}

// GetDatastructValues, 获取 datastructs 包中的 values, 并以 slice 形式返回
// data 参数为要获取 field 的实例指针
func GetDataStructValues(data interface{}) []driver.Value {
	valueOf := reflect.ValueOf(data).Elem()
	numField := reflect.TypeOf(data).Elem().NumField()
	valueSlice := make([]driver.Value, 0, numField)
	for i := 0; i < numField; i++ {
		valueSlice = append(valueSlice, valueOf.Field(i).Interface())
	}
	return valueSlice
}

// GetDatastructDests, 获取 datastructs 包中的存放空间，以 slice 形式返回
// data 参数为要获取存放空间的实例指针
func GetDataStructDests(data interface{}) ([]interface{}, error) {
	typeOf := reflect.TypeOf(data).Elem()
	valueOf := reflect.ValueOf(data).Elem()
	numField := valueOf.NumField()
	dests := make([]interface{}, 0, numField)
	for i := 0; i < numField; i++ {
		tfield := typeOf.Field(i)
		if len(tfield.PkgPath) > 0 {
			return nil, fmt.Errorf("field %s is not public", tfield.Name)
		}
		mysqlFieldName := tfield.Tag.Get("mysqlField")
		if len(mysqlFieldName) == 0 {
			return nil, fmt.Errorf("field %s has no mysqlField tag", tfield.Name)
		}

		vfield := valueOf.Field(i)
		if vfield.CanAddr() {
			dests = append(dests, vfield.Addr().Interface())
		} else {
			return nil, fmt.Errorf("%v can not be addressed", vfield)
		}
	}
	return dests, nil
}
