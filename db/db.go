package database

import (
	"context"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ctxTransactionKey struct {
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

var db *gorm.DB

// 根据项目env初始化
func InitDefault() {
	cfg := env.GetInstance()
	db = InitDataSource(Option{
		DSN:             cfg.GetString(DB_DSN),
		MaxIdleConns:    cfg.GetInt(DB_MAX_IDLE_CONNS),
		MaxOpenConns:    cfg.GetInt(DB_MAX_OPEN_CONNS),
		ConnMaxLifetime: cfg.GetInt(DB_CONN_MAX_LIFE_TIME),
	})
}
func InitDefaultWithOption(option Option) {
	db = InitDataSource(option)
}

func InitDataSource(option Option) *gorm.DB {
	dsn := option.DSN
	if len(dsn) == 0 {
		panic("db init error: dsn is null")
	}
	mysqldb, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
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

func Get() *gorm.DB {
	return db
}

func GetWithContext(ctx context.Context) *gorm.DB {
	odb := ctx.Value(ctxTransactionKey{})
	if odb != nil {
		tx, ok := odb.(*gorm.DB)
		if !ok {
			log.Panicf("unexpect context value type: %s", reflect.TypeOf(tx))
			return nil
		}
		return tx
	}
	return db.WithContext(ctx)
}

func Transaction(ctx context.Context, fc func(txctx context.Context) error) error {
	ndb := db.WithContext(ctx)
	return ndb.Transaction(func(tx *gorm.DB) error {
		txctx := context.WithValue(ctx, ctxTransactionKey{}, tx)
		return fc(txctx)
	})
}

func Expr(expr string, args ...interface{}) clause.Expr {
	return gorm.Expr(expr, args)
}
