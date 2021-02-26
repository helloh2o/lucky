package xdb

import (
	"database/sql"
	"github.com/helloh2o/lucky/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

var (
	db    *gorm.DB
	sqlDB *sql.DB
)

const (
	// MYSQL dbType
	MYSQL = "mysql"
	// PG dbType
	PG = "postgres"
)

// OpenMysqlDB gorm v2 dbUrl = username:password@tcp(localhost:3306)/db_name?charset=utf8mb4&parseTime=True&loc=Local
func OpenMysqlDB(dbType string, dbUrl string, config *gorm.Config, maxIdleConns, maxOpenConns int, models ...interface{}) (instance *gorm.DB, err error) {
	if config == nil {
		config = &gorm.Config{}
	}

	if config.NamingStrategy == nil {
		config.NamingStrategy = schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		}
	}
	switch dbType {
	case PG:
		if db, err = gorm.Open(postgres.Open(dbUrl), config); err != nil {
			log.Error("opens postgres database failed: %s", err.Error())
			return
		}
	case MYSQL:
		if db, err = gorm.Open(mysql.Open(dbUrl), config); err != nil {
			log.Error("opens mysql database failed: %s", err.Error())
			return
		}
	default: // mysql
		if db, err = gorm.Open(mysql.Open(dbUrl), config); err != nil {
			log.Error("opens mysql database failed: %s", err.Error())
			return
		}
	}
	instance = db
	if sqlDB, err = db.DB(); err == nil {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	} else {
		log.Error("%v", err)
	}

	if err = db.AutoMigrate(models...); nil != err {
		log.Error("auto migrate tables failed: %s", err.Error())
	}
	return
}

// DB 获取数据库链接
func DB() *gorm.DB {
	return db
}
