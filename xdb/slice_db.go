package db

import (
	"fmt"
	"github.com/helloh2o/lucky/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"strings"
	"sync"
	"yz-servers/entity"
)

const (
	EmptyVal      = ""
	PartZero      = 0
	PartOnlyOne   = 1
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
	defer func() {
		db.Statement.Table = EmptyVal
	}()
	db = db.Session(&gorm.Session{})
	stmt := &gorm.Statement{
		DB:       db,
		ConnPool: db.Statement.ConnPool,
		Context:  db.Statement.Context,
		Clauses:  map[string]clause.Clause{},
		Vars:     make([]interface{}, 0, 8),
	}
	db.Statement = stmt
	var err error
	if stmt.Schema, err = schema.Parse(sf.Value, qt, stmt.DB.NamingStrategy); err == nil && stmt.Table == "" {
		if tables := strings.Split(stmt.Schema.Table, "."); len(tables) == 2 {
			stmt.TableExpr = &clause.Expr{SQL: stmt.Quote(stmt.Schema.Table)}
			stmt.Table = tables[1]
		}
		stmt.Table = stmt.Schema.Table
	}
	db.Statement.Table = fmt.Sprintf("%s%s", db.Statement.Table, dbIndex)
	if _, ok := qt.Load(db.Statement.Table); !ok {
		if err := db.Migrator().AutoMigrate(sf.Value); err != nil {
			log.Error("migrator error %v", err)
			return nil
		}
		qt.Store(db.Statement.Table, struct{}{})
	}
	return db.Table(db.Statement.Table)
}
