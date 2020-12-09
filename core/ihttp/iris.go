package ihttp

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"io"
	"os"
)

var (
	app = iris.New()
)

func init() {
	app.Logger().SetLevel("debug")
	app.Use(recover.New())
	app.Use(logger.New())
	app.Logger().SetOutput(os.Stdout)
}

func AddHandler(h context.Handler) {
	app.Use(h)
}

func SetLogOutput(w io.Writer) {
	app.Logger().SetOutput(w)
}

func SetLogLv(lv string) {
	app.Logger().SetLevel(lv)
}

func EnableCrossOrigin() {
	app.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		MaxAge:           600,
		AllowedMethods:   []string{iris.MethodGet, iris.MethodPost, iris.MethodOptions, iris.MethodHead, iris.MethodDelete, iris.MethodPut},
		AllowedHeaders:   []string{"*"},
	}))
	app.AllowMethods(iris.MethodOptions)
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
func Delete(path string, handlers ...context.Handler) {
	app.Delete(path, handlers...)
}
func Put(path string, handlers ...context.Handler) {
	app.Put(path, handlers...)
}
