package gpostgres

import (
	"github.com/skirrund/gcloud/database/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDialector struct{}

func (PostgresDialector) New(option option.Option) gorm.Dialector {
	dsn := option.DSN
	if len(dsn) == 0 {
		panic("db init error: dsn is null")
	}
	return postgres.New(postgres.Config{
		DSN: option.DSN,
	})
}
