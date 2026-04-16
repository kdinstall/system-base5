package controllers

import (
	"net/http"
	"webapp/src/lib/docker"
	tmpl "webapp/src/lib/template"

	"github.com/gin-gonic/gin"
)

// ContainerController はDockerコンテナ管理のアクションを提供する
type ContainerController struct{}

// Index はコンテナ一覧を表示する (GET /containers)
func (cc *ContainerController) Index(c *gin.Context) {
	containers, err := docker.ListContainers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "containers.html", tmpl.MergeData(gin.H{
			"page_title":  "コンテナ一覧",
			"active_page": "containers",
			"error":       "コンテナ一覧の取得に失敗しました: " + err.Error(),
			"containers":  nil,
		}))
		return
	}

	c.HTML(http.StatusOK, "containers.html", tmpl.MergeData(gin.H{
		"page_title":  "コンテナ一覧",
		"active_page": "containers",
		"containers":  containers,
		"flash":       c.Query("flash"),
	}))
}

// Start はコンテナを起動する (POST /containers/:id/start)
func (cc *ContainerController) Start(c *gin.Context) {
	containerID := c.Param("id")

	if err := docker.StartContainer(containerID); err != nil {
		c.Redirect(http.StatusSeeOther, "/containers?flash=start_failed")
		return
	}

	c.Redirect(http.StatusSeeOther, "/containers?flash=started")
}

// Stop はコンテナを停止する (POST /containers/:id/stop)
func (cc *ContainerController) Stop(c *gin.Context) {
	containerID := c.Param("id")

	if err := docker.StopContainer(containerID); err != nil {
		c.Redirect(http.StatusSeeOther, "/containers?flash=stop_failed")
		return
	}

	c.Redirect(http.StatusSeeOther, "/containers?flash=stopped")
}

// Restart はコンテナを再起動する (POST /containers/:id/restart)
func (cc *ContainerController) Restart(c *gin.Context) {
	containerID := c.Param("id")

	if err := docker.RestartContainer(containerID); err != nil {
		c.Redirect(http.StatusSeeOther, "/containers?flash=restart_failed")
		return
	}

	c.Redirect(http.StatusSeeOther, "/containers?flash=restarted")
}

// Logs はコンテナのログを表示する (GET /containers/:id/logs)
func (cc *ContainerController) Logs(c *gin.Context) {
	containerID := c.Param("id")

	// コンテナ情報取得
	container, err := docker.GetContainerByID(containerID)
	if err != nil {
		c.HTML(http.StatusNotFound, "404.html", tmpl.MergeData(gin.H{"page_title": "Not Found"}))
		return
	}

	// ログ取得
	logs, err := docker.GetLogs(containerID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "container_logs.html", tmpl.MergeData(gin.H{
			"page_title":  "ログ表示",
			"active_page": "containers",
			"container":   container,
			"error":       "ログの取得に失敗しました: " + err.Error(),
			"logs":        "",
		}))
		return
	}

	c.HTML(http.StatusOK, "container_logs.html", tmpl.MergeData(gin.H{
		"page_title":  "ログ表示 - " + container.Name,
		"active_page": "containers",
		"container":   container,
		"logs":        logs,
	}))
}
