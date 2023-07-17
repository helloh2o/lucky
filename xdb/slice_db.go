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
	// 表
	TBRedCash     = "t_red_cash_change_log"
	TBOrder       = "t_box_order"
	TBSellerOrder = "t_cw_seller_order"
	TBOpenBox     = "t_open_box_record"
	TBBoxItemBK   = "t_box_item_bak"
	TBExchange    = "t_exchange_log"
	TBOpenCard    = "t_open_card_record"
	TBCardItemBK  = "t_card_box_item_bak"
	EmptyVal      = ""
	PartZero      = 0
	PartOnlyOne   = 1
)

var (
	qt = new(sync.Map)
)

// GetSliceDB 获取分表DB
func GetSliceDB(db *gorm.DB, userId int64, tb string) *gorm.DB {
	dbIndex := userId % 20
	if dbIndex < 0 {
		dbIndex = 0
	}
	tbName := fmt.Sprintf("%s%d", tb, dbIndex)
	return db.Table(tbName)
}

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
func hasTable(gb *gorm.DB, tableName string) bool {
	var count int64
	gb.Raw("SELECT count(*) FROM information_schema.tables WHERE table_name = ? AND table_type = ?", tableName, "BASE TABLE").Row().Scan(&count)
	return count > 0
}

// CopyOldData2New 拷贝数据导新表
func CopyOldData2New(userId int64) {
	dbIndex := userId % 20
	if dbIndex == 0 {
		return
	}
}

// GetTables 需要注册的表
func GetTables() []interface{} {
	return []interface{}{
	}
}
