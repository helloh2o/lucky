package xdb

import (
	"fmt"
	"github.com/helloh2o/lucky/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"strings"
	"sync"
)

const EmptyTable = ""

var (
	qt = new(sync.Map)
)

type SP struct {
	value interface{}
	part  int64
}

// GetAutoSliceDB 获取自动分表
func GetAutoSliceDB(db *gorm.DB, userId int64, sf SP) *gorm.DB {
	dbIndex := userId % sf.part
	if dbIndex < 0 {
		dbIndex = 0
	}
	stmt := db.Statement
	stmt.Table = EmptyTable
	var err error
	if stmt.Schema, err = schema.Parse(sf.value, qt, stmt.DB.NamingStrategy); err == nil && stmt.Table == "" {
		if tables := strings.Split(stmt.Schema.Table, "."); len(tables) == 2 {
			stmt.TableExpr = &clause.Expr{SQL: stmt.Quote(stmt.Schema.Table)}
			stmt.Table = tables[1]
		}
		stmt.Table = stmt.Schema.Table
	}
	db.Statement.Table = fmt.Sprintf("%s%d", db.Statement.Table, dbIndex)
	if _, ok := qt.Load(db.Statement.Table); !ok {
		if err := db.Migrator().AutoMigrate(sf.value); err != nil {
			log.Error("migrator error %v", err)
			return nil
		}
	} else {
		qt.Store(db.Statement.Table, struct{}{})
	}
	return db.Table(db.Statement.Table)
}
