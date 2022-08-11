package utils

import (
	"github.com/kataras/iris/v12/context"
	"strings"
)

// GetIrisRemoteAddr 获取Iris 客户端IP
func GetIrisRemoteAddr(ctx *context.Context) string {
	remoteInfo := ctx.GetHeader("X-Forwarded-For")
	if remoteInfo == "" {
		remoteInfo = ctx.RemoteAddr()
	} else {
		// 多层转发 X-Forwarded-For: client, proxy1, proxy2
		forwardList := strings.Split(remoteInfo, ",")
		if len(forwardList) > 1 {
			return forwardList[0]
		}
	}
	return remoteInfo
}
