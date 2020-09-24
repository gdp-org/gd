/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package mysqldb

import (
	"fmt"
	"strings"
	"testing"


	log "github.com/chuck1024/gd/dlog"

	"time"

	"github.com/DATA-DOG/go-sqlmock"
	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDB(t *testing.T) {
	log.SetLevel(int(log.WARNING))

	Convey("Test MysqlClient.queryList", t, func() {
		db1, _, err1 := sqlmock.New()
		db2, _, err2 := sqlmock.New()
		db3, mock3, err3 := sqlmock.New()
		db4, mock4, err4 := sqlmock.New()
		db5, mock5, err5 := sqlmock.New()
		So(err1, ShouldBeNil)
		So(err2, ShouldBeNil)
		So(err3, ShouldBeNil)
		So(err4, ShouldBeNil)
		So(err5, ShouldBeNil)

		mysqlClient := &MysqlClient{}

		dbWrapWrite1 := NewDbWrapped("192.168.1.1", db1, mysqlClient, 1*time.Second)
		dbWrapWrite2 := NewDbWrapped("192.168.1.2", db2, mysqlClient, 1*time.Second)
		dbWrapRead1 := NewDbWrapped("192.168.1.3", db3, mysqlClient, 1*time.Second)
		dbWrapRead2 := NewDbWrapped("192.168.1.4", db4, mysqlClient, 1*time.Second)
		dbWrapRead3 := NewDbWrapped("192.168.1.5", db5, mysqlClient, 1*time.Second)

		mysqlClient.dbWrite = []*DbWrap{dbWrapWrite1, dbWrapWrite2}
		mysqlClient.dbRead = []*DbWrap{dbWrapRead1, dbWrapRead2, dbWrapRead3}

		randProduct := genRandomProduct()
		productFields, err := GetDataStructFields(randProduct)
		So(err, ShouldBeNil)
		productValues := GetDataStructValues(randProduct)

		rows := sqlmock.NewRows(productFields).AddRow(productValues...)
		queryExpr := "select " + strings.Join(productFields, ",") + " from product where pd_id = (.+)"
		query := "select " + strings.Join(productFields, ",") + " from product where pd_id = ?"

		Convey("test retry when some db exception", func() {
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)

			_, err := mysqlClient.queryList(query, randProduct.PdId)

			So(err, ShouldBeNil)
		})

		Convey("test retry when db all exception", func() {
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)

			_, err := mysqlClient.queryList(query, randProduct.PdId)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, mysql.ErrInvalidConn.Error())
		})

		SkipConvey("test retry when some db flowtoken exhausted", func() {
			for i := 0; i < 420; i++ {
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)

				dbWrapRead1.Query(query, randProduct.PdId)
				//	fmt.Printf("err: %v \n", err)
			}

			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)

			_, err := mysqlClient.queryList(query, randProduct.PdId)

			So(err, ShouldBeNil)
		})

		SkipConvey("test retry when all db flowtoken exhausted", func() {
			for i := 0; i < 250; i++ {
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)

				dbWrapRead1.Query(query, randProduct.PdId)
				dbWrapRead2.Query(query, randProduct.PdId)
				dbWrapRead3.Query(query, randProduct.PdId)
				//	fmt.Printf("err: %v \n", err)
			}

			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)

			_, err := mysqlClient.queryList(query, randProduct.PdId)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no available db")
		})
	})

	Convey("Test MysqlClient.queryRow", t, func() {
		db1, _, err1 := sqlmock.New()
		db2, _, err2 := sqlmock.New()
		db3, mock3, err3 := sqlmock.New()
		db4, mock4, err4 := sqlmock.New()
		db5, mock5, err5 := sqlmock.New()
		So(err1, ShouldBeNil)
		So(err2, ShouldBeNil)
		So(err3, ShouldBeNil)
		So(err4, ShouldBeNil)
		So(err5, ShouldBeNil)

		mysqlClient := &MysqlClient{}

		dbWrapWrite1 := NewDbWrapped("192.168.1.1", db1, mysqlClient, 1*time.Second)
		dbWrapWrite2 := NewDbWrapped("192.168.1.2", db2, mysqlClient, 1*time.Second)
		dbWrapRead1 := NewDbWrapped("192.168.1.3", db3, mysqlClient, 1*time.Second)
		dbWrapRead2 := NewDbWrapped("192.168.1.4", db4, mysqlClient, 1*time.Second)
		dbWrapRead3 := NewDbWrapped("192.168.1.5", db5, mysqlClient, 1*time.Second)

		mysqlClient.dbWrite = []*DbWrap{dbWrapWrite1, dbWrapWrite2}
		mysqlClient.dbRead = []*DbWrap{dbWrapRead1, dbWrapRead2, dbWrapRead3}

		randProduct := genRandomProduct()
		productFields, err := GetDataStructFields(randProduct)
		So(err, ShouldBeNil)
		productValues := GetDataStructValues(randProduct)

		rows := sqlmock.NewRows(productFields).AddRow(productValues...)
		queryExpr := "select " + strings.Join(productFields, ",") + " from product where pd_id = (.+)"
		query := "select " + strings.Join(productFields, ",") + " from product where pd_id = ?"

		Convey("test retry when some db exception", func() {
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)

			_, err := mysqlClient.queryList(query, randProduct.PdId)

			So(err, ShouldBeNil)
		})

		Convey("test retry when db all exception", func() {
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)

			_, err := mysqlClient.queryList(query, randProduct.PdId)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, mysql.ErrInvalidConn.Error())
		})

		SkipConvey("test retry when some db flowtoken exhausted", func() {
			for i := 0; i < 420; i++ {
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)

				dbWrapRead1.Query(query, randProduct.PdId)
				//	fmt.Printf("err: %v \n", err)
			}

			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)

			_, err := mysqlClient.queryList(query, randProduct.PdId)

			So(err, ShouldBeNil)
		})

		SkipConvey("test retry when all db flowtoken exhausted", func() {
			for i := 0; i < 420; i++ {
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)
				mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(nil).WillReturnError(mysql.ErrInvalidConn)

				dbWrapRead1.Query(query, randProduct.PdId)
				dbWrapRead2.Query(query, randProduct.PdId)
				dbWrapRead3.Query(query, randProduct.PdId)
				//	fmt.Printf("err: %v \n", err)
			}

			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
			mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)

			_, err := mysqlClient.queryList(query, randProduct.PdId)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no available db")
		})
	})

	Convey("Test MysqlClient.Query", t, func() {
		db1, _, err1 := sqlmock.New()
		db2, _, err2 := sqlmock.New()
		db3, mock3, err3 := sqlmock.New()
		db4, mock4, err4 := sqlmock.New()
		db5, mock5, err5 := sqlmock.New()
		So(err1, ShouldBeNil)
		So(err2, ShouldBeNil)
		So(err3, ShouldBeNil)
		So(err4, ShouldBeNil)
		So(err5, ShouldBeNil)

		mysqlClient := &MysqlClient{}

		dbWrapWrite1 := NewDbWrapped("192.168.1.1", db1, mysqlClient, 1*time.Second)
		dbWrapWrite2 := NewDbWrapped("192.168.1.2", db2, mysqlClient, 1*time.Second)
		dbWrapRead1 := NewDbWrapped("192.168.1.3", db3, mysqlClient, 1*time.Second)
		dbWrapRead2 := NewDbWrapped("192.168.1.4", db4, mysqlClient, 1*time.Second)
		dbWrapRead3 := NewDbWrapped("192.168.1.5", db5, mysqlClient, 1*time.Second)

		mysqlClient.dbWrite = []*DbWrap{dbWrapWrite1, dbWrapWrite2}
		mysqlClient.dbRead = []*DbWrap{dbWrapRead1, dbWrapRead2, dbWrapRead3}

		randProduct := genRandomProduct()
		productFields, err := GetDataStructFields(randProduct)
		So(err, ShouldBeNil)
		productValues := GetDataStructValues(randProduct)

		rows := sqlmock.NewRows(productFields).AddRow(productValues...)
		queryExpr := "select " + strings.Join(productFields, ",") + " from product where pd_id = (.+)"
		query := "select " + strings.Join(productFields, ",") + " from product where pd_id = ?"

		mock3.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
		mock4.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)
		mock5.ExpectQuery(queryExpr).WithArgs(randProduct.PdId).WillReturnRows(rows).WillReturnError(nil)

		ret, err := mysqlClient.Query((*Product)(nil), query, randProduct.PdId)
		So(err, ShouldBeNil)
		product, ok := ret.(*Product)
		fmt.Printf("product: %v \n", product)
		So(ok, ShouldBeTrue)
		So(product.Name, ShouldEqual, randProduct.Name)
	})

	Convey("Test MysqlClient.QueryList", t, func() {
		db1, _, err1 := sqlmock.New()
		db2, _, err2 := sqlmock.New()
		db3, mock3, err3 := sqlmock.New()
		db4, mock4, err4 := sqlmock.New()
		db5, mock5, err5 := sqlmock.New()
		So(err1, ShouldBeNil)
		So(err2, ShouldBeNil)
		So(err3, ShouldBeNil)
		So(err4, ShouldBeNil)
		So(err5, ShouldBeNil)

		mysqlClient := &MysqlClient{}

		dbWrapWrite1 := NewDbWrapped("192.168.1.1", db1, mysqlClient, 1*time.Second)
		dbWrapWrite2 := NewDbWrapped("192.168.1.2", db2, mysqlClient, 1*time.Second)
		dbWrapRead1 := NewDbWrapped("192.168.1.3", db3, mysqlClient, 1*time.Second)
		dbWrapRead2 := NewDbWrapped("192.168.1.4", db4, mysqlClient, 1*time.Second)
		dbWrapRead3 := NewDbWrapped("192.168.1.5", db5, mysqlClient, 1*time.Second)

		mysqlClient.dbWrite = []*DbWrap{dbWrapWrite1, dbWrapWrite2}
		mysqlClient.dbRead = []*DbWrap{dbWrapRead1, dbWrapRead2, dbWrapRead3}

		randProduct1 := genRandomProduct()
		productFields, err := GetDataStructFields(randProduct1)
		So(err, ShouldBeNil)
		productValues1 := GetDataStructValues(randProduct1)

		randProduct2 := genRandomProduct()
		_, err = GetDataStructFields(randProduct2)
		So(err, ShouldBeNil)
		productValues2 := GetDataStructValues(randProduct2)

		rows := sqlmock.NewRows(productFields).AddRow(productValues1...)
		rows.AddRow(productValues2...)

		queryExpr := "select " + strings.Join(productFields, ",") + " from product where pd_id = (.+)"
		query := "select " + strings.Join(productFields, ",") + " from product where pd_id = ?"

		mock3.ExpectQuery(queryExpr).WithArgs(randProduct1.PdId).WillReturnRows(rows).WillReturnError(nil)
		mock4.ExpectQuery(queryExpr).WithArgs(randProduct1.PdId).WillReturnRows(rows).WillReturnError(nil)
		mock5.ExpectQuery(queryExpr).WithArgs(randProduct1.PdId).WillReturnRows(rows).WillReturnError(nil)

		retList, err := mysqlClient.QueryList((*Product)(nil), query, randProduct1.PdId)
		So(err, ShouldBeNil)

		productList := make([]*Product, 0)
		for _, ret := range retList {
			product, _ := ret.(*Product)
			productList = append(productList, product)
		}

		So(len(productList), ShouldEqual, 2)
		expectPdIds := []int64{randProduct1.PdId, randProduct2.PdId}
		for _, product := range productList {
			So(product.PdId, ShouldBeIn, expectPdIds)
		}
	})

	Convey("Test MysqlClient.GetCount", t, func() {
		db1, _, err1 := sqlmock.New()
		db2, _, err2 := sqlmock.New()
		db3, mock3, err3 := sqlmock.New()
		db4, mock4, err4 := sqlmock.New()
		db5, mock5, err5 := sqlmock.New()
		So(err1, ShouldBeNil)
		So(err2, ShouldBeNil)
		So(err3, ShouldBeNil)
		So(err4, ShouldBeNil)
		So(err5, ShouldBeNil)

		mysqlClient := &MysqlClient{}

		dbWrapWrite1 := NewDbWrapped("192.168.1.1", db1, mysqlClient, 1*time.Second)
		dbWrapWrite2 := NewDbWrapped("192.168.1.2", db2, mysqlClient, 1*time.Second)
		dbWrapRead1 := NewDbWrapped("192.168.1.3", db3, mysqlClient, 1*time.Second)
		dbWrapRead2 := NewDbWrapped("192.168.1.4", db4, mysqlClient, 1*time.Second)
		dbWrapRead3 := NewDbWrapped("192.168.1.5", db5, mysqlClient, 1*time.Second)

		mysqlClient.dbWrite = []*DbWrap{dbWrapWrite1, dbWrapWrite2}
		mysqlClient.dbRead = []*DbWrap{dbWrapRead1, dbWrapRead2, dbWrapRead3}

		queryExpr := "select total from product where pd_id = (.+)"
		query := "select total from product where pd_id = ?"

		pdId := 100
		total := 2
		rows := sqlmock.NewRows([]string{"total"}).AddRow(total)

		mock3.ExpectQuery(queryExpr).WithArgs(pdId).WillReturnRows(rows).WillReturnError(nil)
		mock4.ExpectQuery(queryExpr).WithArgs(pdId).WillReturnRows(rows).WillReturnError(nil)
		mock5.ExpectQuery(queryExpr).WithArgs(pdId).WillReturnRows(rows).WillReturnError(nil)

		ret, err := mysqlClient.GetCount(query, pdId)
		So(err, ShouldBeNil)
		So(ret, ShouldEqual, total)
	})

}

type Product struct {
	PdId          int64  `mysqlField:"pd_id"`
	PtId          int64  `mysqlField:"pt_id"`
	ProducterId   int64  `mysqlField:"producter_id"`
	Name          string `mysqlField:"name"`
	ShortName     string `mysqlField:"short_name"`
	Intro         string `mysqlField:"intro"`
	Img           string `mysqlField:"img"`
	Version       string `mysqlField:"version"`
	ReferId       int64  `mysqlField:"refer_id"`
	UpdateFile    string `mysqlField:"update_file"`
	McuFile       string `mysqlField:"mcu_file"`
	PluginId      int64  `mysqlField:"plugin_id"`
	CloudId       int64  `mysqlField:"cloud_id"`
	IosCompatible int64  `mysqlField:"ios_compatible"`
	Ctime         int64  `mysqlField:"ctime"`
	Alias         string `mysqlField:"alias"`
	ShowMode      int64  `mysqlField:"show_mode"`
	ConnectType   int64  `mysqlField:"connect_type"`
}

func genRandomProduct() *Product {
	return &Product{
		PdId:          int64(randomdata.Number(1000, 10000)),
		PtId:          int64(randomdata.Number(1000, 10000)),
		ProducterId:   int64(randomdata.Number(1000, 10000)),
		Name:          randomdata.Letters(32),
		ShortName:     randomdata.Letters(10),
		Intro:         randomdata.Letters(100),
		Img:           randomdata.Letters(20),
		Version:       randomdata.Letters(10),
		ReferId:       int64(randomdata.Number(1000, 10000)),
		UpdateFile:    randomdata.Letters(20),
		McuFile:       randomdata.Letters(20),
		PluginId:      int64(randomdata.Number(1000, 10000)),
		CloudId:       int64(randomdata.Number(1000, 10000)),
		IosCompatible: int64(randomdata.Number(1000, 10000)),
		Ctime:         int64(randomdata.Number(1000000000, 2000000000)),
		Alias:         randomdata.Letters(10),
		ShowMode:      int64(randomdata.Number(1000, 10000)),
		ConnectType:   int64(randomdata.Number(10)),
	}
}

