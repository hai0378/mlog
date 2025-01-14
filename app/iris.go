package app

import (
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mlogclub/mlog/controllers/api"
	"github.com/mlogclub/mlog/utils/config"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/mvc"
	"github.com/mlogclub/simple"
	"github.com/sirupsen/logrus"

	"github.com/mlogclub/mlog/controllers/admin"
	"github.com/mlogclub/mlog/middleware"
)

func InitIris() {
	app := iris.New()
	app.Logger().SetLevel("warn")
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

	app.OnAnyErrorCode(func(ctx context.Context) {
		path := ctx.Path()
		var err error
		if strings.Contains(path, "/api/admin/") {
			_, err = ctx.JSON(simple.JsonErrorCode(ctx.GetStatusCode(), "Http error"))
		}
		if err != nil {
			logrus.Error(err)
		}
	})

	// api
	mvc.Configure(app.Party("/api"), func(m *mvc.Application) {
		m.Party("/topic").Handle(new(api.TopicController))
		m.Party("/article").Handle(new(api.ArticleController))
		m.Party("/project").Handle(new(api.ProjectController))
		m.Party("/login").Handle(new(api.LoginController))
		m.Party("/user").Handle(new(api.UserController))
		m.Party("/tag").Handle(new(api.TagController))
		m.Party("/category").Handle(new(api.CategoryController))
		m.Party("/comment").Handle(new(api.CommentController))
		m.Party("/favorite").Handle(new(api.FavoriteController))
		m.Party("/config").Handle(new(api.ConfigController))
		m.Party("/upload").Handle(new(api.UploadController))
	})

	// admin
	mvc.Configure(app.Party("/api/admin"), func(m *mvc.Application) {
		m.Router.Use(middleware.AdminAuth)
		m.Party("/user").Handle(new(admin.UserController))
		m.Party("/github-user").Handle(new(admin.GithubUserController))
		m.Party("/category").Handle(new(admin.CategoryController))
		m.Party("/tag").Handle(new(admin.TagController))
		m.Party("/article").Handle(new(admin.ArticleController))
		m.Party("/comment").Handle(new(admin.CommentController))
		m.Party("/favorite").Handle(new(admin.FavoriteController))
		m.Party("/article-tag").Handle(new(admin.ArticleTagController))
		m.Party("/topic").Handle(new(admin.TopicController))
		m.Party("/sys-config").Handle(new(admin.SysConfigController))
	})

	server := &http.Server{Addr: ":" + config.Conf.Port}
	handleSignal(server)
	err := app.Run(iris.Server(server), iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		DisableInterruptHandler:           false,
		DisablePathCorrection:             false,
		EnablePathEscape:                  false,
		FireMethodNotAllowed:              false,
		DisableBodyConsumptionOnUnmarshal: false,
		DisableAutoFireStatusCode:         false,
		EnableOptimizations:               true,
		TimeFormat:                        "2006-01-02 15:04:05",
		Charset:                           "UTF-8",
	}))
	if err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
}

func handleSignal(server *http.Server) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		s := <-c
		logrus.Infof("got signal [%s], exiting now", s)
		if err := server.Close(); nil != err {
			logrus.Errorf("server close failed: " + err.Error())
		}

		simple.CloseDB()

		logrus.Infof("Exited")
		os.Exit(0)
	}()
}
