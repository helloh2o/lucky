package app

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"os"
)

var (
	app = iris.New()
)

func init() {
	app.Logger().SetLevel("debug")
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)
	//// 同时写文件日志与控制台日志
	//app.Logger().SetOutput(io.MultiWriter(f, os.Stdout))
	//// or 使用下面这个
	//// 日志只生成到文件
	app.Logger().SetOutput(os.Stdout)
}

func Run(addr string) error {
	return app.Listen(addr)
}

func Get(path string, handlers ...context.Handler) {
	app.Get(path, handlers...)
}
func Post(path string, handlers ...context.Handler) {
	app.Post(path, handlers...)
}
