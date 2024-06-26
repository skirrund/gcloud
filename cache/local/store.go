package local

import (
	"errors"
	"math"
	"time"

	"github.com/skirrund/gcloud/logger"

	"github.com/maypok86/otter"
)

// var cache *freecache.Cache
var cache otter.CacheWithVariableTTL[string, any]

func init() {
	// cacheSize := 200 * 1024 * 1024
	// cache = freecache.NewCache(cacheSize)
	var err error
	cache, err = otter.MustBuilder[string, any](math.MaxUint16).WithVariableTTL().Build()
	if err != nil {
		panic(err)
	}
}

func NewCache(capacity int) (otter.CacheWithVariableTTL[string, any], error) {
	return otter.MustBuilder[string, any](capacity).WithVariableTTL().Build()
}

func Get(key string) any {
	val, ex := cache.Get(key)
	logger.Info("[localCache] get cache :" + key)
	if !ex {
		return nil
	}
	return val
}

func Del(key string) {
	cache.Delete(key)
}

func Set(key string, value any, expireSeconds int64) error {
	if value == nil {
		return nil
	}
	// valueBytes, err := serialize(value)
	logger.Info("[localCache] cache :" + key)
	// if err != nil {
	// 	logger.Error("[localCache] error:" + err.Error())
	// 	return err
	// }
	r := cache.Set(key, value, time.Duration(expireSeconds)*time.Second)
	if !r {
		err := errors.New("the key-value pair had too much cost and the Set was dropped")
		logger.Error("[localCache] error:", err.Error())
		return err
	}
	return nil
}

// func serialize(value interface{}) ([]byte, error) {
// 	buf := bytes.Buffer{}
// 	enc := gob.NewEncoder(&buf)
// 	gob.Register(value)

// 	err := enc.Encode(&value)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return buf.Bytes(), nil
// }

// func deserialize(valueBytes []byte) (any, error) {
// 	var value interface{}
// 	buf := bytes.NewBuffer(valueBytes)
// 	dec := gob.NewDecoder(buf)

// 	err := dec.Decode(&value)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return value, nil
// }
