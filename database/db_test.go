package database

import (
	"context"
	"testing"
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
	option := Option{
		DSN: "test_root:I#9Qvnyg@tcp(cpg3xayuo60t5jni4q8lb9rzmwkcf1se.mysql.qingcloud.link:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
	}
	InitDataSource(option)
	t.Log("db init finished")
	ctx := context.Background()
	Transaction(ctx, func(txctx context.Context) error {
		tt1 := make(map[string]interface{})
		tt1["name"] = "tt1"
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
		return nil
	})
}
