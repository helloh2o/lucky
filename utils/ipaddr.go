package utils

import "github.com/kataras/iris/v12/context"

// GetIrisRemoteAddr 获取Iris 客户端IP
func GetIrisRemoteAddr(ctx *context.Context) string {
	remoteInfo := ctx.GetHeader("X-Forwarded-For")
	if remoteInfo == "" {
		remoteInfo = ctx.RemoteAddr()
	}
	return remoteInfo
}
