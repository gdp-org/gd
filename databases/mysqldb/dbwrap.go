/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package mysqldb

import (
	"context"
	"database/sql"
	"errors"
	log "github.com/gdp-org/gd/dlog"
	"github.com/gdp-org/gd/runtime/gl"
	"github.com/gdp-org/gd/runtime/pc"
	"reflect"
	"strings"
	"time"

	"runtime"

	_ "github.com/go-sql-driver/mysql"
)

const (
	defaultDbRetry = 1

	PcMysqlReadFail     = "mysql_read_fail"
	PcMysqlReadAllFail  = "mysql_read_all_fail"
	PcMysqlWriteAllFail = "mysql_write_all_fail"
	PcMysqlWriteFail    = "mysql_write_fail"
	PcMysqlRead         = "mysql_read"
	PcMysqlWrite        = "mysql_write"
	PcMysqlTransaction  = "mysql_transaction"

	glDBReadCost             = "db_read_cost"
	glDBWriteCost            = "db_write_cost"
	glDBTransactionCost      = "db_transaction_cost"
	glDBReadCount            = "db_read_count"
	glDBWriteCount           = "db_write_count"
	glDBTransactionCount     = "db_transaction_count"
	glDBReadFailCount        = "db_read_fail_count"
	glDBWriteFailCount       = "db_write_fail_count"
	glDBTransactionFailCount = "db_transaction_fail_count"
)

type DbWrap struct {
	Timeout     time.Duration
	mysqlClient *MysqlClient
	host        string
	*sql.DB
	glSuffix string

	retry int
}

func NewDbWrapped(host string, db *sql.DB, mysqlClient *MysqlClient, timeout time.Duration) *DbWrap {
	return NewDbWrappedRetry(host, db, mysqlClient, timeout, defaultDbRetry)
}

func NewDbWrappedRetry(host string, db *sql.DB, mysqlClient *MysqlClient, timeout time.Duration, retry int) *DbWrap {
	return NewDbWrappedRetryProxy(host, db, mysqlClient, timeout, retry, false)
}

func NewDbWrappedRetryProxy(host string, db *sql.DB, mysqlClient *MysqlClient, timeout time.Duration, retry int, proxy bool) *DbWrap {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	w := &DbWrap{
		mysqlClient: mysqlClient,
		host:        host,
		DB:          db,
		Timeout:     timeout,
		retry:       retry,
	}
	return w
}

func (db *DbWrap) pcDbReadAllFail() string {
	return PcMysqlReadAllFail + ",db=" + db.host
}

func (db *DbWrap) pcDbRead() string {
	return PcMysqlRead + ",db=" + db.host
}

func (db *DbWrap) pcDbWrite() string {
	return PcMysqlWrite + ",db=" + db.host
}

func (db *DbWrap) pcDbTransaction() string {
	return PcMysqlTransaction + ",db=" + db.host
}

func (db *DbWrap) glDbReadFail() string {
	if db.glSuffix == "" {
		return glDBReadFailCount
	} else {
		return glDBReadFailCount + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbReadCount() string {
	if db.glSuffix == "" {
		return glDBReadCount
	} else {
		return glDBReadCount + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbReadCost() string {
	if db.glSuffix == "" {
		return glDBReadCost
	} else {
		return glDBReadCost + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbWriteFail() string {
	if db.glSuffix == "" {
		return glDBWriteFailCount
	} else {
		return glDBWriteFailCount + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbTransactionFail() string {
	if db.glSuffix == "" {
		return glDBTransactionFailCount
	} else {
		return glDBTransactionFailCount + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbWriteCount() string {
	if db.glSuffix == "" {
		return glDBWriteCount
	} else {
		return glDBWriteCount + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbTransactionCount() string {
	if db.glSuffix == "" {
		return glDBTransactionCount
	} else {
		return glDBTransactionCount + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbWriteCost() string {
	if db.glSuffix == "" {
		return glDBWriteCost
	} else {
		return glDBWriteCost + "_" + db.glSuffix
	}
}

func (db *DbWrap) glDbTransactionCost() string {
	if db.glSuffix == "" {
		return glDBTransactionCost
	} else {
		return glDBTransactionCost + "_" + db.glSuffix
	}
}

func (db *DbWrap) Query(query string, args ...interface{}) (rs *sql.Rows, err error) {
	retry := db.retry
	if retry < 0 {
		retry = 0
	}
	turn := 0
	for turn <= retry {
		turn++
		rs, err = db.doQuery(query, args...)
		if err != nil {
			//only retry on connection error
			if IsTimeoutError(err) || IsDbConnError(err) {
				continue
			} else {
				errMsg := err.Error()
				if strings.Contains(errMsg, "getsockopt") {
					errt := reflect.TypeOf(err)
					log.Error("SOCKOPT_FAIL", "query err,type=%v,err=%v", errt, err)
					continue
				}

				break
			}
		} else {
			break
		}
	}

	if err != nil {
		pc.Incr(db.pcDbReadAllFail(), 1)
	}
	return
}

func (db *DbWrap) doQuery(query string, args ...interface{}) (rs *sql.Rows, err error) {
	st := time.Now()
	pcKey := db.pcDbRead()

	defer func() {
		cost := time.Now().Sub(st)
		pc.Cost(pcKey, cost)
		gl.Incr(db.glDbReadCost(), int64(cost/time.Millisecond))
		if err == nil && cost > time.Duration(1)*time.Second {
			log.Debug("MYSQL_SLOW_QUERY", "query=%s,cost=%d,host=%s,err=%v", query, cost/time.Millisecond, db.host, err)
		}

		if err != nil {
			pc.CostFail(pcKey, 1)
			gl.Incr(db.glDbReadFail(), 1)
		}
	}()

	gl.Incr(db.glDbReadCount(), 1)
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	rs, err = db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		if cancel != nil {
			defer cancel()
		}
	}

	return
}

func (db *DbWrap) QueryRow(query string, args ...interface{}) *Row {
	rows, err := db.Query(query, args...)
	return &Row{rows: rows, err: err}
}

func (db *DbWrap) Exec(query string, args ...interface{}) (r sql.Result, err error) {
	return db.ExecContext(nil, query, args...)
}

func (db *DbWrap) ExecContext(ctx context.Context, query string, args ...interface{}) (r sql.Result, err error) {
	targetDb := db
	st := time.Now()
	pcKey := db.pcDbWrite()
	defer func() {
		cost := time.Now().Sub(st)
		pc.Cost(pcKey, cost)
		gl.Incr(db.glDbWriteCost(), int64(cost/time.Millisecond))
		if err == nil && cost > time.Duration(1)*time.Second {
			log.Debug("MYSQL_SLOW_QUERY", "query=%s,cost=%d,host=%s,err=%v", query, cost/time.Millisecond, targetDb.host, err)
		}

		if err != nil {
			gl.Incr(db.glDbWriteFail(), 1)
			pc.CostFail(pcKey, 1)
		}
	}()

	gl.Incr(db.glDbWriteCount(), 1)
	if ctx == nil {
		ct, cancel := context.WithTimeout(context.Background(), db.Timeout)
		ctx = ct
		defer cancel()
	}
	r, err = targetDb.DB.ExecContext(ctx, query, args...)
	return
}

type TransactionExec func(ctx context.Context, tx *sql.Tx) (sql.Result, error)

func (db *DbWrap) ExecTransaction(transactionExec TransactionExec) (r sql.Result, err error) {
	targetDb := db
	pcKey := db.pcDbTransaction()
	st := time.Now()
	defer func() {
		cost := time.Now().Sub(st)
		pc.Cost(pcKey, cost)
		gl.Incr(db.glDbTransactionCost(), int64(cost/time.Millisecond))
		if err == nil && cost > time.Duration(1)*time.Second {
			log.Debug("MYSQL_SLOW_QUERY", "transaction=%v,cost=%d,host=%s", getFunctionName(transactionExec), cost/time.Millisecond, targetDb.host)
		}

		if err != nil {
			pc.CostFail(pcKey, 1)
			gl.Incr(db.glDbTransactionFail(), 1)
		}
	}()

	gl.Incr(db.glDbTransactionCount(), 1)
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	var tx *sql.Tx
	tx, err = targetDb.DB.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	r, err = transactionExec(ctx, tx)

	return
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (db *DbWrap) Close() error {
	return db.DB.Close()
}

// Row is the result of calling QueryRow to select a single row.
type Row struct {
	// One of these two will be non-nil:
	err  error // deferred error for easy chaining
	rows *sql.Rows
}

// Scan copies the columns from the matched row into the values
// pointed at by dest. See the documentation on Rows.Scan for details.
// If more than one row matches the query,
// Scan uses the first row and discards the rest. If no row matches
// the query, Scan returns ErrNoRows.
func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	defer r.rows.Close()
	for _, dp := range dest {
		if _, ok := dp.(*sql.RawBytes); ok {
			return errors.New("sql: RawBytes isn't allowed on Row.Scan")
		}
	}

	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	err := r.rows.Scan(dest...)
	if err != nil {
		return err
	}
	// Make sure the query can be processed to completion with no errors.
	if err := r.rows.Close(); err != nil {
		return err
	}

	return nil
}
