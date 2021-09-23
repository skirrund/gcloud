package utils

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/skirrund/gcloud/cache/redis"
	"github.com/skirrund/gcloud/logger"
)

const (
	workerBits uint8 = 5  // 每台机器(节点)的id位数 5位最大可以有2^5
	numberBits uint8 = 17 // 表示每个集群下的每个节点，1秒内可生成的id序号的二进制位数 即每秒可生成 2^17-1个唯一id
	// 这里求最大值使用了位运算
	workerMax   int64 = ^(-1 << workerBits)     // 节点ID的最大值，用于防止溢出
	numberMax   int64 = ^(-1 << numberBits)     // 每个节点，1秒内可生成的id序号最大值
	timeShift   uint8 = workerBits + numberBits // 时间戳向左的偏移量
	workerShift uint8 = numberBits              // 节点id向左的偏移量
	// 31位字节作为时间戳数值的话 大约68年就会用完
	// 假如你2010年1月1日开始开发系统 如果不减去2010年1月1日的时间戳 那么白白浪费40年的时间戳啊！
	// 这个一旦定义且开始生成ID后千万不要改了 不然可能会生成相同的ID
	epoch int64 = 1586631844 //起始的时间戳
)

// 定义一个woker工作节点所需要的基本参数
type Worker struct {
	mu          sync.Mutex // 添加互斥锁 确保并发安全
	timestamp   int64      // 记录时间戳
	workerId    int64      // 该节点的ID
	number      int64      // 当前毫秒已经生成的id序列号(从0开始累加)
	redisClient *redis.RedisClient
}

var redisClient *redis.RedisClient

func NewWorkerWithRedis(r *redis.RedisClient, appName string) (*Worker, error) {
	redisClient = r
	workId := r.HIncrBy("datacenterId", appName, 1) % 32
	logger.Infof("[IdWorker] init datacenterId:%d", workId)
	return newWorker(workId)
}

/**
@desc 初始化一个节点
@auth jerry.shi 2021-05-24
 */
func newWorker(workerId int64) (*Worker, error) {
	// 要先检测workerId是否在上面定义的范围内
	if workerId < 0 || workerId > workerMax {
		return nil, errors.New("workId is invalidate")
	}
	return &Worker{
		timestamp: 0,
		workerId:  workerId,
		number:    0,
	}, nil
}

func (w *Worker) GetIdWithPrefix(prefix string) string {
	id := w.GetId()
	return prefix + id
}

/**
@desc 获取id
@auth jerry.shi 2021-05-24
 */
func (w *Worker) GetId() string {
	//解决并发安全
	w.mu.Lock()
	defer w.mu.Unlock()
	//获取生成时的时间戳
	now := w.Now()
	if now == w.timestamp {
		w.number++
		//这里要判断，当前工作节点是否在1秒内已经生成numberMax个id
		if w.number > numberMax {
			//如果当前工作节点在1秒内生成的id已经超过上限 需要等待1秒再继续生成
			w.number = 0
			for now <= w.timestamp {
				now = w.Now()
			}
			w.timestamp = now
		}
	} else if now > w.timestamp {
		//如果当前时间与工作节点上一次生成id的时间不一致 则需要重置工作节点生成id的序号
		w.number = 0
		//将机器上一次生成id的时间更新为当前时间
		w.timestamp = now

	} else {
		for now < w.timestamp {
			now = w.Now()
		}
		w.number++
		//这里要判断，当前工作节点是否在1秒内已经生成numberMax个id
		if w.number > numberMax {
			//如果当前工作节点在1秒内生成的id已经超过上限 需要等待1秒再继续生成
			w.number = 0
			for now <= w.timestamp {
				now = w.Now()
			}
			w.timestamp = now
		}
	}
	//第一段 now - epoch 为该算法目前已经奔跑了秒
	//如果在程序跑了一段时间修改了epoch这个值 可能会导致生成相同的id
	id := int64((now-epoch)<<timeShift | (w.workerId << workerShift) | (w.number))
	return strconv.FormatInt(id, 10)
}

/**
@desc 获取当前时间
@auth liuguoqiang 2020-06-16
@param
@return
 */
func (w *Worker) Now() int64 {
	return time.Now().Unix()
}
