package api

import (
	"github.com/kataras/iris/v12/context"
	"lucky-day/app"
)

// 验证用户数据
func verify() {
	app.Post("/verify", func(ctx *context.Context) {

	})
}
