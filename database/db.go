package database

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/skirrund/gcloud/bootstrap/env"
	"github.com/skirrund/gcloud/database/option"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ctxTransactionKey struct {
}

type Dialector interface {
	New(option option.Option) gorm.Dialector
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
func InitDefault(dialector Dialector) {
	cfg := env.GetInstance()
	db = InitDataSource(option.Option{
		DSN:             cfg.GetString(DB_DSN),
		MaxIdleConns:    cfg.GetInt(DB_MAX_IDLE_CONNS),
		MaxOpenConns:    cfg.GetInt(DB_MAX_OPEN_CONNS),
		ConnMaxLifetime: cfg.GetInt(DB_CONN_MAX_LIFE_TIME),
		QueryFields:     cfg.GetBool(DB_QueryFields),
		Type:            cfg.GetStringWithDefault(DB_TYPE, DB_TYPE_MYSQL),
	}, dialector)
}
func InitDefaultWithOption(option option.Option, dialector Dialector) {
	db = InitDataSource(option, dialector)
}

func doInit(option option.Option, dialector gorm.Dialector) *gorm.DB {
	if len(option.Type) == 0 {
		option.Type = DB_TYPE_MYSQL
	}
	gormdb, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		QueryFields:            option.QueryFields,
		TranslateError:         true,
	})
	if err != nil {
		log.Panicln(err)
	}
	sqlDB, err := gormdb.DB()
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
	if option.Type == DB_TYPE_MYSQL {
		sqlDB.Exec("set @@session.sql_mode=(SELECT CONCAT(@@session.sql_mode,',STRICT_TRANS_TABLES'))")
	}
	return gormdb
}

func InitDataSource(option option.Option, dialector Dialector) *gorm.DB {
	dsn := option.DSN
	if len(dsn) == 0 {
		panic("db init error: dsn is null")
	}
	if dialector == nil {
		panic("db init error: dialector is null")
	}
	gormDia := dialector.New(option)
	return doInit(option, gormDia)
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

func IsDuplicatedKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
