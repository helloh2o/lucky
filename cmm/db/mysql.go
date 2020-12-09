package db

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"lucky/log"
)

var (
	db    *gorm.DB
	sqlDB *sql.DB
)

// gorm v2 dbUrl = username:password@tcp(localhost:3306)/db_name?charset=utf8mb4&parseTime=True&loc=Local
func OpenMysqlDB(dbUrl string, config *gorm.Config, maxIdleConns, maxOpenConns int, models ...interface{}) (err error) {
	if config == nil {
		config = &gorm.Config{}
	}

	if config.NamingStrategy == nil {
		config.NamingStrategy = schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		}
	}

	if db, err = gorm.Open(mysql.Open(dbUrl), config); err != nil {
		log.Error("opens database failed: %s", err.Error())
		return
	}

	if sqlDB, err = db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
	} else {
		log.Error("%v", err)
	}

	if err = db.AutoMigrate(models...); nil != err {
		log.Error("auto migrate tables failed: %s", err.Error())
	}
	return
}

// 获取数据库链接
func Mysql() *gorm.DB {
	return db
}
