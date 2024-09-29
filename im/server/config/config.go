package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type Conf struct {
	// Server uuid
	ServerId string `yaml:"server_id" json:"server_id"`
	// Redis 连接地址
	RedisUrl string `yaml:"redis_url" json:"redis_url"`
	// 哨兵 下使用的DB
	RedisSDBIndex int `yaml:"redis_sdb_index"  json:"redis_sdb_index"`
	// 主服务别名
	RedisMasterName string `yaml:"redis_master_name"   json:"redis_master_name"`
	// Redis 哨兵群组
	RedisSentinelGroup []string `yaml:"redis_sentinel_group" json:"redis_sentinel_group"`
	// ETCD 集群组
	ETCDClusterList []string `yaml:"etcd_cluster_list" json:"etcd_cluster_list"`
	// nats queue url
	NatsUrl string `yaml:"nats_url" json:"nats_url"`
	// 服务监听地址
	ListenAddr string `yaml:"listen_addr" json:"listen_addr"`
	// 日志级别
	LogLevel string `yaml:"log_level" json:"log_level"`
	LogFile  string `yaml:"log_file" json:"log_file"`
	// pprof
	Pprof bool `yaml:"pprof" json:"pprof"`
}

var (
	c     = new(Conf)
	cPath string
	cLock sync.RWMutex
)

type ConfServer struct {
	Id     uint32 `yaml:"id"`     // 网关ID（llk=1）
	Type   uint32 `yaml:"type"`   // 类型 (1聊天，2游戏)
	Host   string `yaml:"host"`   // 目标服务器
	Port   int32  `yaml:"port"`   // 目标端口
	Scheme string `yaml:"scheme"` // 连接协议 (ws,wss,http,https,tcp,udp,quic)
	Proto  string `yaml:"proto"`  // 消息协议 (json & protobuf)
}

// Initialize 初始化配置
func Initialize(path string) *Conf {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(c)
	if err != nil {
		panic(err)
	}
	cPath = path
	return c
}

// Get 获取配置
func Get() *Conf {
	cLock.RLock()
	rt := c
	defer cLock.RUnlock()
	return rt
}
