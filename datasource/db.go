package db

import (
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dataSource struct {
	DB    *gorm.DB
	SqlDB *sql.DB
}

type Option struct {
	DSN string
	//打开数据库连接的最大数量，默认DefaultMaxOpenConns
	MaxOpenConns int
	//空闲连接池中连接的最大数量，默认DefaultMaxIdleConns
	MaxIdleConns int
	//连接可复用的最大时间。单位分钟，默认DefaultConnMaxLifetime
	ConnMaxLifetime int
}

const (
	DB_DSN                 = "datasource.dsn"
	DB_URL_KEY             = "datasource.url"
	DB_USERNAME_KEY        = "datasource.username"
	DB_PWD_KEY             = "datasource.password"
	DB_CONN_MAX_LIFE_TIME  = "datasource.connMaxLifetime"
	DB_MAX_IDLE_CONNS      = "datasource.maxIdleConns"
	DB_MAX_OPEN_CONNS      = "datasource.maxOpenConns"
	DefaultConnMaxLifetime = 30 * time.Minute
	DefaultMaxIdleConns    = 10
	DefaultMaxOpenConns    = 50
)

var instance *dataSource
var once sync.Once

func InitDataSource(option Option) {
	dsn := option.DSN
	if len(dsn) == 0 {
		panic("db init error: dsn is null")
	}

	once.Do(func() {
		//username:password@tcp(test-fast-mysql.meditrusthealth.com:3306)/fast_config?charset=utf8mb4&parseTime=True&loc=Local

		mysqldb, err := gorm.Open(mysql.New(mysql.Config{
			DSN: dsn,
		}), &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		})

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
		instance = &dataSource{
			DB: mysqldb,
		}
	})
}

func CreateInsertSql(tableName string, kv map[string]interface{}) (sql string, values []interface{}) {
	values = make([]interface{}, len(kv))
	sql = "insert into " + tableName + "("
	vsql := " values("
	i := 0
	for k, v := range kv {
		k = strings.ReplaceAll(k, ";", "")
		sql += k + ","
		vsql += "?,"
		values[i] = v
		i++
	}
	sql = sql[:len(sql)-1] + ")"
	vsql = vsql[:len(vsql)-1] + ")"
	return sql + vsql, values
}

func Get() *dataSource {
	return instance
}

func Expr(expr string, args ...interface{}) clause.Expr {
	return gorm.Expr(expr, args)
}
