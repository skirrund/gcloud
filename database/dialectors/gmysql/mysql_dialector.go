package gmysql

import (
	mysql1 "github.com/go-sql-driver/mysql"
	"github.com/skirrund/gcloud/database/option"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlDialector struct{}

func (MysqlDialector) New(option option.Option) gorm.Dialector {
	dsn := option.DSN
	if len(dsn) == 0 {
		panic("db init error: dsn is null")
	}
	return mysql.New(mysql.Config{
		DSN: option.DSN,
	})
}

func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if merr, ok := err.(*mysql1.MySQLError); ok {
		return merr.Number == 1062
	}
	return false
}
