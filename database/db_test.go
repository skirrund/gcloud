package database

import (
	"context"
	"errors"
	"testing"

	"github.com/skirrund/gcloud/database/dialectors/gmysql"
	"github.com/skirrund/gcloud/database/option"
)

type Test1Service struct {
}
type Test2Service struct {
}

func (Test1Service) saveTest1(ctx context.Context, data interface{}) error {
	db := GetWithContext(ctx)
	db.Table("test_tx1").Create(data)
	return db.Error
}

func (Test2Service) saveTest2(ctx context.Context, data interface{}) error {
	db := GetWithContext(ctx)
	db.Table("test_tx2").Create(data)
	return db.Error
}

func TestTransaction(t *testing.T) {
	option := option.Option{
		DSN: "",
	}
	InitDefaultWithOption(option, new(gmysql.MysqlDialector))
	t.Log("db init finished")
	ctx := context.Background()
	Transaction(ctx, func(txctx context.Context) error {
		tt1 := make(map[string]interface{})
		tt1["name"] = "tt11"
		ts1 := Test1Service{}
		err := ts1.saveTest1(txctx, tt1)
		if err != nil {
			return err
		}
		tt2 := make(map[string]interface{})
		tt2["name"] = "tt2"
		ts2 := Test2Service{}
		err = ts2.saveTest2(txctx, tt2)
		if err != nil {
			return err
		}
		return errors.New("test")
	})
}
