package xdb

import (
	"fmt"
	"github.com/helloh2o/lucky/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
)

const (
	EmptyVal    = ""
	PartOnlyOne = 1
)

var (
	qt = new(sync.Map)
)

type SP struct {
	Value  interface{}
	Part   int64  // 用户ID取余分表
	ByDate string // 通过日期2022-01-01分表，按月，按日
}

// GetAutoSliceDB 获取自动分表
func GetAutoSliceDB(db *gorm.DB, userId int64, sf SP) *gorm.DB {
	if sf.Value == nil {
		return nil
	}
	dbIndex := EmptyVal
	switch {
	case sf.Part <= PartOnlyOne:
		dbIndex = EmptyVal
	case sf.Part > PartOnlyOne:
		dbIndex = fmt.Sprintf("%d", userId%sf.Part)
	}
	if sf.ByDate != EmptyVal {
		dbIndex += "_" + sf.ByDate
	}
	// table key
	tk := reflect.TypeOf(sf.Value).String() + dbIndex
	tableName, ok := qt.Load(tk)
	if !ok {
		if tableName = autoMigrateTable(dbIndex, sf.Value); tableName != nil {
			qt.Store(tk, tableName)
		}
	}
	log.Debug("slice table name:%v, key:%s", tableName, tk)
	return db.Table(tableName.(string))
}

// autoMigrateTable 自动迁移表, 在使用之前完成
func autoMigrateTable(dbIndex, ett interface{}) interface{} {
	tdb := QpsDB()
	var err error
	if tdb.Statement.Schema, err = schema.Parse(ett, qt, tdb.NamingStrategy); err == nil && tdb.Statement.Table == "" {
		if tables := strings.Split(tdb.Statement.Schema.Table, "."); len(tables) == 2 {
			tdb.Statement.TableExpr = &clause.Expr{SQL: tdb.Statement.Quote(tdb.Statement.Schema.Table)}
			tdb.Statement.Table = tables[1]
		}
		target := fmt.Sprintf("%s%s", tdb.Statement.Schema.Table, dbIndex)
		if err = tdb.Table(target).Migrator().AutoMigrate(ett); err != nil {
			log.Error("migrator error %v", err)
		}
		return target
	}
	return nil
}
