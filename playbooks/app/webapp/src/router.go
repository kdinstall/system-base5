package main

import (
	"webapp/src/controllers"
	tmpl "webapp/src/lib/template"

	"github.com/gin-gonic/gin"
)

func initRouter() *gin.Engine {
	router := gin.Default()

	// テンプレートロード
	t, err := tmpl.LoadTemplates("src/templates")
	if err != nil {
		panic("テンプレートのロードに失敗しました: " + err.Error())
	}
	router.SetHTMLTemplate(t)

	// 静的ファイル
	router.Static("/assets", "public/assets")

	// ルータ登録
	registerContainerRouter(router)
	registerInstallRouter(router)

	// 404 ハンドラ
	router.NoRoute(func(c *gin.Context) {
		c.HTML(404, "404.html", tmpl.MergeData(gin.H{"page_title": "Not Found"}))
	})

	return router
}

func registerContainerRouter(router *gin.Engine) {
	cc := &controllers.ContainerController{}

	// リダイレクト
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/containers")
	})

	// コンテナ管理
	router.GET("/containers", cc.Index)
	router.POST("/containers/:id/start", cc.Start)
	router.POST("/containers/:id/stop", cc.Stop)
	router.POST("/containers/:id/restart", cc.Restart)
	router.GET("/containers/:id/logs", cc.Logs)
}

func registerInstallRouter(router *gin.Engine) {
	ic := &controllers.InstallController{}

	// インストール管理
	router.GET("/install", ic.Index)
	router.GET("/install/:name/config", ic.Config)
	router.POST("/install/execute", ic.Execute)
}
