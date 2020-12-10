package inet

import (
	"lucky/core/iduck"
	"time"
)

// 帧同步房间
type FrameGR struct {
	// 网络连接
	Connections []iduck.IConnection
	// 进入令牌
	EnterToken string
	// 同步周期
	FrameTicker time.Ticker
	// current frame messages
	CurrentFrame []interface{}
}

func NewFrameGR() {

}

func AddConn() {

}

func DelConn() {

}
func Destroy() {

}
