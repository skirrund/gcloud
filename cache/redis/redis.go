package redis

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/skirrund/gcloud/logger"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client redis.UniversalClient
}

type Options struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addrs []string `property:"redis.addrs"`

	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int `property:"redis.db"`

	// Common options.

	Dialer    func(ctx context.Context, network, addr string) (net.Conn, error)
	OnConnect func(ctx context.Context, cn *redis.Conn) error

	Username         string `property:"redis.username"`
	Password         string `property:"redis.password"`
	SentinelPassword string `property:"redis.sentinel.password"`

	MaxRetries      int           `property:"redis.maxRetries"`
	MinRetryBackoff time.Duration `property:"redis.minRetryBackoff"`
	MaxRetryBackoff time.Duration `property:"redis.maxRetryBackoff"`

	DialTimeout  time.Duration `property:"redis.dialTimeout"`
	ReadTimeout  time.Duration `property:"redis.readTimeout"`
	WriteTimeout time.Duration `property:"redis.writeTimeout"`

	PoolSize           int           `property:"redis.poolSize"`
	MinIdleConns       int           `property:"redis.minIdleConns"`
	MaxConnAge         time.Duration `property:"redis.maxConnAge"`
	PoolTimeout        time.Duration `property:"redis.poolTimeout"`
	IdleTimeout        time.Duration `property:"redis.idleTimeout"`
	IdleCheckFrequency time.Duration `property:"redis.idleCheckFrequency"`

	TLSConfig *tls.Config

	// Only cluster clients.

	MaxRedirects   int  `property:"redis.maxRedirects"`
	ReadOnly       bool `property:"redis.readOnly"`
	RouteByLatency bool `property:"redis.routeByLatency"`
	RouteRandomly  bool `property:"redis.routeRandomly"`

	// The sentinel master name.
	// Only failover clients.

	MasterName string `property:"redis.masterName"`
}

var ctx = context.Background()

var redisClient *RedisClient

var once sync.Once

func GetClient() *RedisClient {
	return redisClient
}

func NewClient(opts Options) *RedisClient {
	once.Do(func() {
		redisClient = &RedisClient{}
		logger.Info("[redis] init client:", opts)
		rdb := redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:              opts.Addrs,
			Username:           opts.Username,
			Password:           opts.Password,
			DB:                 opts.DB,
			Dialer:             opts.Dialer,
			OnConnect:          opts.OnConnect,
			SentinelPassword:   opts.SentinelPassword,
			MaxRetries:         opts.MaxRetries,
			MinRetryBackoff:    opts.MinRetryBackoff,
			MaxRetryBackoff:    opts.MaxRetryBackoff,
			DialTimeout:        opts.DialTimeout,
			ReadTimeout:        opts.ReadTimeout,
			WriteTimeout:       opts.WriteTimeout,
			PoolSize:           opts.PoolSize,
			MinIdleConns:       opts.MinIdleConns,
			MaxConnAge:         opts.MaxConnAge,
			PoolTimeout:        opts.PoolTimeout,
			IdleTimeout:        opts.IdleTimeout,
			IdleCheckFrequency: opts.IdleCheckFrequency,
			TLSConfig:          opts.TLSConfig,
			MaxRedirects:       opts.MaxRedirects,
			ReadOnly:           opts.ReadOnly,
			RouteByLatency:     opts.RouteByLatency,
			RouteRandomly:      opts.RouteRandomly,
			MasterName:         opts.MasterName,
		})
		redisClient.client = rdb
		err := redisClient.Ping()
		if err != nil {
			logger.Info("[redis] ping error:", err.Error())
		} else {
			logger.Info("[redis] ping success")
		}
	})
	return redisClient
}

func (r *RedisClient) Close() {
	logger.Info("[redis] close redis-client")
	err := r.client.Close()
	if err != nil {
		logger.Error("[redis] close error", err)
	}
}

func (r *RedisClient) Ping() error {
	sc := r.client.Ping(ctx)
	if sc.Err() != nil {
		logger.Error("[redis] err", sc.Err().Error())
		return sc.Err()
	}
	return nil
}

func (r *RedisClient) Get(key string) string {
	sc := r.client.Get(ctx, key)
	return sc.Val()
}

func (r *RedisClient) Size() int64 {
	return r.client.DBSize(ctx).Val()
}

func (r *RedisClient) Set(key string, value string, expiration time.Duration) {
	r.client.Set(ctx, key, value, expiration)
}

func (r *RedisClient) SetNX(key string, value string, expiration time.Duration) bool {
	bc := r.client.SetNX(ctx, key, value, expiration)
	return bc.Val()
}

func (r *RedisClient) HashKey(key string) bool {
	bc := r.client.Exists(ctx, key)
	return bc.Val() == 1
}

func (r *RedisClient) Expire(key string, expiration time.Duration) bool {
	bc := r.client.Expire(ctx, key, expiration)
	return bc.Val()
}

func (r *RedisClient) Del(keys ...string) int64 {
	bc := r.client.Del(ctx, keys...)
	return bc.Val()
}

func (r *RedisClient) HSet(key string, hashKey string, val string) int64 {
	bc := r.client.HSet(ctx, key, hashKey, val)
	return bc.Val()
}

func (r *RedisClient) HMSet(key string, vals map[string]string) bool {
	values := make([]string, 0)
	for k, v := range vals {
		values = append(values, k)
		values = append(values, v)
	}
	bc := r.client.HMSet(ctx, key, values)
	return bc.Val()
}

func (r *RedisClient) HGet(key string, hashKey string) string {
	bc := r.client.HGet(ctx, key, hashKey)
	return bc.Val()
}

func (r *RedisClient) HGetAll(key string) map[string]string {
	bc := r.client.HGetAll(ctx, key)
	return bc.Val()
}
func (r *RedisClient) HMGet(key string, hashKey ...string) map[string]interface{} {
	vals := make(map[string]interface{})
	if len(hashKey) == 0 {
		return vals
	}
	bc := r.client.HMGet(ctx, key, hashKey...)
	v := bc.Val()
	i := 0
	for _, k := range hashKey {
		vals[k] = v[i]
		i++
	}
	return vals
}
func (r *RedisClient) HDel(key string, hashKey ...string) int64 {
	bc := r.client.HDel(ctx, key, hashKey...)
	return bc.Val()
}

func (r *RedisClient) HIncrBy(key string, hashKey string, delta int64) int64 {
	bc := r.client.HIncrBy(ctx, key, hashKey, delta)
	return bc.Val()
}

func (r *RedisClient) Ttl(key string) time.Duration {
	bc := r.client.TTL(ctx, key)
	return bc.Val()
}

func (r *RedisClient) Incr(key string) int64 {
	bc := r.client.Incr(ctx, key)
	return bc.Val()
}

func (r *RedisClient) IncrByTime(key string, expiration time.Duration) int64 {
	bc := r.client.Incr(ctx, key)
	r.Expire(key, expiration)
	return bc.Val()
}

func (r *RedisClient) Decr(key string) int64 {
	bc := r.client.Decr(ctx, key)
	return bc.Val()
}

func (r *RedisClient) ZScore(key string, member string) float64 {
	bc := r.client.ZScore(ctx, key, member)
	return bc.Val()
}
func (r *RedisClient) ZIncrBy(key string, member string, score float64) float64 {
	bc := r.client.ZIncrBy(ctx, key, score, member)
	return bc.Val()
}
func (r *RedisClient) ZAdd(key string, member string, score float64) int64 {
	bc := r.client.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	})
	return bc.Val()
}
func (r *RedisClient) SRemove(key string, member string) int64 {
	bc := r.client.SRem(ctx, key, member)
	return bc.Val()
}

func (r *RedisClient) SIsMember(key string, member string) bool {
	bc := r.client.SIsMember(ctx, key, member)
	return bc.Val()
}

func (r *RedisClient) SMembers(key string) []string {
	bc := r.client.SMembers(ctx, key)
	return bc.Val()
}

func (r *RedisClient) SAdd(key string, member ...interface{}) int64 {
	bc := r.client.SAdd(ctx, key, member...)
	return bc.Val()
}

func (r *RedisClient) LPop(key string, member ...string) string {
	bc := r.client.LPop(ctx, key)
	return bc.Val()
}

func (r *RedisClient) LPush(key string, valus ...interface{}) int64 {
	bc := r.client.LPush(ctx, key, valus...)
	return bc.Val()
}
