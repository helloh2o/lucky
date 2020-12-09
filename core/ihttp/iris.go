package ihttp

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

func Run(addr, logLevel string) error {
	app.Logger().SetLevel(logLevel)
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
	app.Logger().SetOutput(os.Stdout)
	return app.Listen(addr)
}

func Get(path string, handlers ...context.Handler) {
	app.Get(path, handlers...)
}
func Post(path string, handlers ...context.Handler) {
	app.Post(path, handlers...)
}
