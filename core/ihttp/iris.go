package ihttp

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
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

// Handle add router
func Handle(method string, relativePath string, handlers ...context.Handler) {
	app.Handle(method, relativePath, handlers...)
}

// HandleDir static dir
func HandleDir(requestPath string, fsOrDir interface{}, opts ...iris.DirOptions) (routes []*router.Route) {
	return app.HandleDir(requestPath, fsOrDir, opts...)
}

// Iris return the instance
func Iris() *iris.Application {
	return app
}

// AddHandler for iris
func AddHandler(h context.Handler) {
	app.Use(h)
}

// SetLogOutput for io
func SetLogOutput(w io.Writer) {
	app.Logger().SetOutput(w)
}

// SetLogLv for iris
func SetLogLv(lv string) {
	app.Logger().SetLevel(lv)
}

// EnableCrossOrigin for client
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

// Run the iris app
func Run(addr string) error {
	return app.Listen(addr)
}

// Get add get router
func Get(path string, handlers ...context.Handler) {
	app.Get(path, handlers...)
}

// Post add post router
func Post(path string, handlers ...context.Handler) {
	app.Post(path, handlers...)
}

// Delete add del router
func Delete(path string, handlers ...context.Handler) {
	app.Delete(path, handlers...)
}

// Put add put router
func Put(path string, handlers ...context.Handler) {
	app.Put(path, handlers...)
}
