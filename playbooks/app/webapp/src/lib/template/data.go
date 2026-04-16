package tmpl

import (
	"time"
	"webapp/src/config"

	"github.com/gin-gonic/gin"
)

// BaseData は全テンプレートで共通して使用するデータを返す
func BaseData() gin.H {
	return gin.H{
		"app_name": config.GetEnv().AppName,
		"g_year":   time.Now().Year(),
	}
}

// MergeData はページ固有データに共通データをマージして返す
// ページ側のキーが優先される
func MergeData(data gin.H) gin.H {
	base := BaseData()
	for k, v := range base {
		if _, ok := data[k]; !ok {
			data[k] = v
		}
	}
	return data
}
