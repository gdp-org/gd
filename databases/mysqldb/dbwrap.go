/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package mysqldb

import (
	"context"
	"database/sql"
	"errors"
	log "github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/runtime/pc"
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
)

type DbWrap struct {
	Timeout     time.Duration
	mysqlClient *MysqlClient
	host        string
	*sql.DB
	ctxSuffix string

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
		if err == nil && cost > time.Duration(1)*time.Second {
			log.Debug("MYSQL_SLOW_QUERY", "query=%s,cost=%d,host=%s,err=%v", query, cost/time.Millisecond, db.host, err)
		}

		if err != nil {
			pc.CostFail(pcKey, 1)
		}
	}()

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
		if err == nil && cost > time.Duration(1)*time.Second {
			log.Debug("MYSQL_SLOW_QUERY", "query=%s,cost=%d,host=%s,err=%v", query, cost/time.Millisecond, targetDb.host, err)
		}

		if err != nil {
			pc.CostFail(pcKey, 1)
		}
	}()

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
		if err == nil && cost > time.Duration(1)*time.Second {
			log.Debug("MYSQL_SLOW_QUERY", "transaction=%v,cost=%d,host=%s", getFunctionName(transactionExec), cost / time.Millisecond, targetDb.host)
		}

		if err != nil {
			pc.CostFail(pcKey, 1)
		}
	}()

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

	// TODO(bradfitz): for now we need to defensively clone all
	// []byte that the driver returned (not permitting
	// *RawBytes in Rows.Scan), since we're about to close
	// the Rows in our defer, when we return from this function.
	// the contract with the driver.Next(...) interface is that it
	// can return slices into read-only temporary memory that's
	// only valid until the next Scan/Close.  But the TODO is that
	// for a lot of drivers, this copy will be unnecessary.  We
	// should provide an optional interface for drivers to
	// implement to say, "don't worry, the []bytes that I return
	// from Next will not be modified again." (for instance, if
	// they were obtained from the network anyway) But for now we
	// don't care.
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
