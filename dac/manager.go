package dac

// data access cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"math/rand"
	"reflect"
	"saas/log"
	"sync"
	"time"

	"saas/consts"
	"saas/utils"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	EMPTY   = ""
	FivePos = 5
	JsonTag = "json:"
)

var (
	defaultOne     sync.Once
	defaultManager *Manager
)

type Manager struct {
	mu                sync.Mutex
	memoryCache       *cache.Cache
	redisClient       redis.UniversalClient
	kafkaConsumer     sarama.Consumer
	kafkaProducer     sarama.SyncProducer
	ctx               context.Context
	cancel            context.CancelFunc
	sfGroup           singleflight.Group
	defaultExpiration time.Duration // 设置默认过期时间
	GroupKeys         map[string]map[string]interface{}
	ExpiredKeys       sync.Map
}

type InvalidationMessage struct {
	Key    string `json:"key"`
	Action string `json:"action"` // delete/invalidate
}

func Default() *Manager {
	if defaultManager == nil {
		panic(errors.New("default manager not initialized"))
	}
	return defaultManager
}

func NewCacheManager(redisClient redis.UniversalClient) (*Manager, error) {
	// 初始化内存缓存（默认1小时过期，10分钟清理间隔）
	mc := cache.New(time.Hour, 10*time.Minute)

	// 测试Redis连接
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("cache manager init got err:%v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{
		memoryCache: mc,
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
		GroupKeys:   make(map[string]map[string]interface{}),
	}
	manager.StartRedisSub()
	defaultOne.Do(func() {
		defaultManager = manager
		InitData()
	})
	return manager, nil
}

// 判断复杂类型
func isComplexType(value interface{}) (reflect.Kind, bool) {
	kind := reflect.TypeOf(value).Kind()
	return kind, kind == reflect.Struct ||
		kind == reflect.Map ||
		kind == reflect.Slice ||
		kind == reflect.Array ||
		kind == reflect.Ptr && reflect.ValueOf(value).Elem().Kind() == reflect.Struct
}

// RandPlusMinutes 随机时间分钟上下浮动
func RandPlusMinutes(exp time.Duration) time.Duration {
	return exp + time.Minute*time.Duration(rand.Intn(30))
}

// Set 方法实现（支持自定义过期时间）
func (cm *Manager) Set(key string, value any, expiration time.Duration) error {
	// 检查value类型
	k, cpx := isComplexType(value)
	// 检查value值
	if value == nil || (k != reflect.Struct && reflect.ValueOf(value).IsNil()) {
		return fmt.Errorf("cache set un expected val:%+v", value)
	}

	expiration = RandPlusMinutes(expiration)

	// 类型检测和序列化处理
	var finalValue string
	switch v := value.(type) {
	case string:
		finalValue = v
	case []byte:
		finalValue = string(v)
	default:
		// 判断是否需要JSON序列化
		if cpx {
			data, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("JSON序列化失败: %w", err)
			}
			finalValue = "json:" + string(data)
		} else {
			finalValue = fmt.Sprintf("%v", value)
		}
	}

	if err := cm.redisClient.Set(cm.ctx, key, finalValue, expiration).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

func (cm *Manager) getAny(key string, missCall func() (any, error), safe bool, container ...any) (any, error) {
	// 从内存缓存获取
	if objBytes, found := cm.memoryCache.Get(key); found {
		if len(container) > 0 {
			cc := container[0]
			if err := json.Unmarshal(objBytes.([]byte), cc); err == nil {
				log.Debug("get mem cache hit for key: %s", key)
				return cc, nil
			}
		}
	}
	// 缓存击穿保护
	get := func() (interface{}, error) {
		if safe {
			done := utils.OpLockTimeout(key, time.Second)
			defer done()
		}
		missRetry := 0
	MissCallOk:
		// 内存中没有，从Redis获取
		val, err := cm.redisClient.Get(cm.ctx, key).Result()
		if errors.Is(err, redis.Nil) || val == EMPTY {
			data, errCall := missCall()
			if errCall != nil {
				return nil, errCall
			}
			// set
			err = cm.Set(key, data, RandPlusMinutes(time.Hour))
			if err == nil && missRetry == 0 {
				// 触发MissCall且缓存成功 => 返回缓存后的结果
				missRetry++
				goto MissCallOk
			}
			return nil, err
		} else if err != nil {
			log.Error("Redis get error: %v", err)
			return nil, err
		}
		log.Debug("get reids cache hit for key: %s", key)
		// 新增反序列化处理
		if len(container) > 0 && len(val) > FivePos {
		JSON:
			if val[:FivePos] == JsonTag {
				objBytes := []byte(val[FivePos:])
				if err = json.Unmarshal(objBytes, container[0]); err != nil {
					log.Error("data:%s unmarshal to container %v error: %v", val, reflect.TypeOf(container[0]), err)
					return nil, err
				} else {
					// 将值存入内存缓存（带过期时间）
					cm.memoryCache.Set(key, objBytes, cache.DefaultExpiration)
					return container[0], nil
				}
			} else {
				// 数据错误
				cm.redisClient.Del(cm.ctx, key)
				err = errors.New(fmt.Sprintf("cache data:%s can't be unmarshaled to:%v", val, reflect.TypeOf(container[0])))
				if json.Valid([]byte(val)) {
					val = JsonTag + val
					log.Error("repair error: %v", err)
					goto JSON
				} else {
					return nil, err
				}
			}
		}
		return val, nil
	}
	if safe {
		val, err, _ := cm.sfGroup.Do(key, func() (interface{}, error) {
			return get()
		})
		return val, err
	} else {
		return get()
	}
}

// Get 获取缓存数据，获得redis数据时container[0]进行数据结构unmarshal, container 为指针类型
func (cm *Manager) Get(key string, missCall func() (any, error), container ...any) (any, error) {
	return cm.getAny(key, missCall, true, container...)
}

// GetUnsafe 获取缓存数据，获得redis数据时container[0]进行数据结构unmarshal, container 为指针类型
func (cm *Manager) GetUnsafe(key string, missCall func() (any, error), container ...any) (any, error) {
	return cm.getAny(key, missCall, false, container...)
}

// GetExpired 获取缓存带有强制过期时间(s)
func (cm *Manager) GetExpired(key string, expiredAt int64, missCall func() (any, error), container ...any) (any, error) {
	v, e := cm.getAny(key, missCall, true, container...)
	cm.AddExpiredKey(key, expiredAt)
	return v, e
}

// AddExpiredKey 添加需要强制过期的KEY
func (cm *Manager) AddExpiredKey(key string, expiredAt int64) {
	left := expiredAt - time.Now().Unix()
	if left > 0 {
		if _, ex := cm.ExpiredKeys.Load(key); ex {
			return
		} else {
			cm.ExpiredKeys.Store(key, expiredAt)
		}
		time.AfterFunc(time.Second*time.Duration(left), func() {
			_ = cm.PubDeleteEvent(key)
		})
	}
}

// Preload 新增缓存预热方法
func (cm *Manager) Preload(keys []string, loadFunc func(string) (string, error)) {
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			if val, err := loadFunc(k); err == nil {
				cm.memoryCache.Set(k, val, cache.DefaultExpiration)
			}
		}(key)
	}
	wg.Wait()
}

func (cm *Manager) Keys() []string {
	items := cm.memoryCache.Items()
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	return keys
}

func (cm *Manager) KeyCount() int {
	return cm.memoryCache.ItemCount()
}

func (cm *Manager) InitKafkaProducer(kafkaAddr string) {
	if cm.kafkaProducer != nil {
		return
	}
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Retry.Max = 3
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Timeout = time.Second * 10

	producer, err := sarama.NewSyncProducer([]string{kafkaAddr}, producerConfig)
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}
	cm.kafkaProducer = producer
}

// PubDeleteEvent 向队列广播删除缓存事件
func (cm *Manager) PubDeleteEvent(keys ...string) error {
	for _, key := range keys {
		if err := cm.redisClient.Publish(context.Background(), consts.CacheChanged, key).Err(); err != nil {
			log.Error("pub change evt to redis channel err:%v \n", err)
		} else {
			log.Debug("pub key %s cache changed", key)
		}
		if cm.kafkaProducer == nil {
			continue
		}
		_, err, _ := cm.sfGroup.Do(key, func() (interface{}, error) {
			// 构造生产者消息
			msg := &sarama.ProducerMessage{
				Topic: consts.KafkaCacheChangeTopic,
				Key:   sarama.StringEncoder(key),
				Value: sarama.ByteEncoder(key),
			}
			// 发送消息（带重试）
			partition, offset, err := cm.kafkaProducer.SendMessage(msg)
			if err != nil {
				return nil, fmt.Errorf("failed to send message: %w (partition: %d, offset: %d)", err, partition, offset)
			} else {
				return nil, nil
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteLocal 删除当前节点缓存
func (cm *Manager) DeleteLocal(keys ...string) {
	for _, key := range keys {
		done := utils.OpLockTimeout(key, time.Second)
		cm.memoryCache.Delete(key)
		cm.redisClient.Del(cm.ctx, key)
		done()
	}
}

func (cm *Manager) StartRedisSub() {
	ctx := context.Background()
	// 默认Redis订阅其它节点释放信息
	subscribe := func() {
		// 订阅解锁频道
		psb := cm.redisClient.Subscribe(ctx, consts.CacheChanged)
		// 确认订阅成功
		if _, err := psb.Receive(ctx); err != nil {
			log.Error("subscribe err:", err)
			return
		}
		// 通过通道接收消息
		ch := psb.Channel()
		for msg := range ch {
			log.Info("received msg from %s msg=%s\n", msg.Channel, msg.Payload)
			mainKey := msg.Payload
			keys := cm.GetGroupCacheKeys(mainKey)
			cm.DeleteLocal(keys...)
		}
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("subscribe panic err:%v", err)
			}
		}()
		// subscribe forever
		for {
			log.Info("subscribe cache change start.")
			subscribe()
			log.Warn("subscribe cache change end.")
			time.Sleep(time.Second)
		}
	}()
}

func (cm *Manager) StartKafkaConsumer(kafkaAddr string) {
	if cm.kafkaConsumer != nil {
		return
	}
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	// 每个客户端都需要消费
	config.ClientID = uuid.NewString()
	consumer, err := sarama.NewConsumer([]string{kafkaAddr}, config)
	if err != nil {
		log.Fatalf("failed to create Kafka consumer: %v", err)
		return
	}
	cm.kafkaConsumer = consumer
	partitionList, err := cm.kafkaConsumer.Partitions(consts.KafkaCacheChangeTopic)
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
		return
	}

	for _, partition := range partitionList {
		pc, err := cm.kafkaConsumer.ConsumePartition(consts.KafkaCacheChangeTopic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Error("Failed to start consumer for partition %d: %v", partition, err)
			continue
		}

		go func(pc sarama.PartitionConsumer) {
			defer func() {
				if r := recover(); r != nil {
					log.Error("cache kafka consumer from panic: %v", r)
				}
				pc.Close()
			}()
			for {
				select {
				case msg := <-pc.Messages():
					mainKey := string(msg.Value)
					keys := cm.GetGroupCacheKeys(mainKey)
					cm.DeleteLocal(keys...)
				case err := <-pc.Errors():
					log.Error("Kafka consumer error: %v", err)
				case <-cm.ctx.Done():
					return
				}
			}
		}(pc)
	}
}

// CollectMetrics 新增监控指标收集
func (cm *Manager) CollectMetrics() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-cm.ctx.Done():
				return
			}
		}
	}()
}

func (cm *Manager) Close() {
	cm.cancel()
	_ = cm.kafkaConsumer.Close()
}

func (cm *Manager) StoreGroupKey(name, SubKey string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.GroupKeys != nil {
		if sub, ok := cm.GroupKeys[name]; ok {
			sub[SubKey] = struct{}{}
		} else {
			cm.GroupKeys[name] = make(map[string]interface{})
			cm.GroupKeys[name][SubKey] = struct{}{}
		}
	}
}

func (cm *Manager) GetGroupCacheKeys(group string) []string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if m, has := cm.GroupKeys[group]; has {
		ret := make([]string, 0, len(m))
		for k, _ := range m {
			ret = append(ret, k)
		}
		return ret
	}
	// return self
	return []string{group}
}
