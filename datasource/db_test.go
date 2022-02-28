package db

import (
	"context"
	"errors"
	"testing"
)

func TestTransaction(t *testing.T) {
	option := Option{
		DSN: "test_root:I#9Qvnyg@tcp(cpg3xayuo60t5jni4q8lb9rzmwkcf1se.mysql.qingcloud.link:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
	}
	InitDataSource(option)
	t.Log("db init finished")
	ctx := context.Background()
	Transaction(ctx, func(txctx context.Context) error {
		db := GetWithContext(txctx)
		tt1 := make(map[string]interface{})
		tt1["name"] = "tt1"
		tt2 := make(map[string]interface{})
		tt2["name"] = "tt2"
		tx := db.Table("test_tx1")
		tx.Create(tt1)
		t.Log(tx.Error)
		tx = tx.Table("test_tx2")
		tx.Create(tt2)
		t.Log(tx.Error)
		return errors.New("rollback")
	})
}
