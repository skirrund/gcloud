package database

import (
	"context"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"
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
	//QueryFields executes the SQL query with all fields of the table
	QueryFields bool
	//数据源类型：MySql
	Type string
}

const (
	DB_DSN                 = "datasource.dsn"
	DB_URL_KEY             = "datasource.url"
	DB_USERNAME_KEY        = "datasource.username"
	DB_PWD_KEY             = "datasource.password"
	DB_CONN_MAX_LIFE_TIME  = "datasource.connMaxLifetime"
	DB_MAX_IDLE_CONNS      = "datasource.maxIdleConns"
	DB_MAX_OPEN_CONNS      = "datasource.maxOpenConns"
	DB_QueryFields         = "datasource.queryFields"
	DB_TYPE                = "datasource.type"
	DefaultConnMaxLifetime = 30 * time.Minute
	DefaultMaxIdleConns    = 10
	DefaultMaxOpenConns    = 50
	DB_TYPE_MYSQL          = "mysql"
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
		QueryFields:     cfg.GetBool(DB_QueryFields),
		Type:            cfg.GetStringWithDefault(DB_TYPE, DB_TYPE_MYSQL),
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
	dbType := strings.ToLower(option.Type)
	switch dbType {
	case DB_TYPE_MYSQL:
		return initMysql(option)
	default:
		return initMysql(option)
	}

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
	var gdb *gorm.DB
	if odb != nil {
		if tx, ok := odb.(*gorm.DB); ok {
			gdb = tx
		} else {
			log.Panicf("unexpect context value type: %s", reflect.TypeOf(tx))
			return nil
		}
	} else {
		gdb = db.WithContext(ctx)
	}
	return gdb
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
