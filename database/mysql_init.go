package database

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initMysql(option Option) *gorm.DB {
	dsn := option.DSN
	if len(dsn) == 0 {
		panic("db init error: dsn is null")
	}
	mysqldb, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		QueryFields:            option.QueryFields,
	})
	if err != nil {
		log.Panicln(err)
	}
	sqlDB, err := mysqldb.DB()
	if err != nil {
		log.Panicln(err)
	} else {
		maxIdleConns := option.MaxIdleConns
		if maxIdleConns == 0 {
			sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
		} else {
			sqlDB.SetMaxIdleConns(maxIdleConns)
		}
		maxOpenConns := option.MaxOpenConns
		if maxOpenConns == 0 {
			sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
		} else {
			sqlDB.SetMaxOpenConns(maxOpenConns)
		}
		connMaxLifetime := option.ConnMaxLifetime
		if connMaxLifetime == 0 {
			sqlDB.SetConnMaxLifetime(DefaultConnMaxLifetime)
		} else {
			sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
		}
	}
	sqlDB.Exec("set @@session.sql_mode=(SELECT CONCAT(@@session.sql_mode,',STRICT_TRANS_TABLES'))")
	return mysqldb
}
