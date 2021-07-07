package local

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/skirrund/gcloud/logger"

	"github.com/coocood/freecache"
)

var cache *freecache.Cache
var once sync.Once

func init() {
	once.Do(func() {
		cacheSize := 200 * 1024 * 1024
		cache = freecache.NewCache(cacheSize)
	})
}

func Get(key string) interface{} {
	valueBytes, err := cache.Get([]byte(key))
	logger.Info("[localCache] get cache :" + key)
	if err != nil {
		logger.Warn("[localCache] error:" + err.Error())
		return nil
	}
	value, err := deserialize(valueBytes)
	if err != nil {
		logger.Warn("[localCache] error:" + err.Error())
		return nil
	}
	return value
}

func Del(key string) {
	cache.Del([]byte(key))
}

func Set(key string, value interface{}, expireSeconds int) error {
	if value == nil {
		return nil
	}
	valueBytes, err := serialize(value)
	logger.Info("[localCache] cache :" + key)
	if err != nil {
		logger.Error("[localCache] error:" + err.Error())
		return err
	}
	return cache.Set([]byte(key), valueBytes, expireSeconds)
}

func serialize(value interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	gob.Register(value)

	err := enc.Encode(&value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func deserialize(valueBytes []byte) (interface{}, error) {
	var value interface{}
	buf := bytes.NewBuffer(valueBytes)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&value)
	if err != nil {
		return nil, err
	}

	return value, nil
}
