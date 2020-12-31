package conf

import "C"
import "sync"

var (
	// C is the config
	C    *Data
	once sync.Once
)

func init() {
	C = &Data{
		ConnUndoQueueSize:   100,
		ConnWriteQueueSize:  10,
		FirstPackageTimeout: 5,
		ConnReadTimeout:     35,
		ConnWriteTimeout:    5,
		MaxDataPackageSize:  4096,
		MaxHeaderLen:        1024,
	}
}

// Set this before startup server
func Set(cfg *Data) {
	once.Do(func() {
		C = cfg
		if C.ConnUndoQueueSize == 0 {
			C.ConnUndoQueueSize = 1
		}
		if C.ConnWriteQueueSize == 0 {
			C.ConnWriteQueueSize = 1
		}
	})
}

// Data is the config struct
type Data struct {
	// 单个连接未处理消息包缓存队列大小
	// 注意：[超过这个大小，包将丢弃，视为当前系统无法处理，默认100]
	ConnUndoQueueSize int
	// 单个连接未写入消息包队列大小 [超过这个大小，包将丢弃，视为当前系统无法处理，默认为1]
	ConnWriteQueueSize int
	// 第一个包等待超市时间 (s) [默认5秒，连接上来未读到正确包，断开连接]
	FirstPackageTimeout int
	// 连接读取超时(s) [默认35秒, 超时等待时间内，请发送任何数据包，如心跳包]
	ConnReadTimeout int
	// 连接写超时(s) [默认5秒, 超时等待时间内，请发送任何数据包，如心跳包]
	ConnWriteTimeout int
	// 数据包最大限制，[默认2048]
	MaxDataPackageSize int
	// ws 最大header，[默认1024]
	MaxHeaderLen int
}
