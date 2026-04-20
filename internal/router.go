package internal

import (
	"net/http"
	"sitchi/internal/role"
	"os"
	"path/filepath"
	"sitchi/internal/user"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	userController := user.NewController()
	roleController := role.NewController()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "sitchi",
			"message": "你好，我是司契！",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})
	webuiApiGroup := r.Group("/api/webui")
	{
		webuiApiGroup.POST("/login", userController.Login)

		// 角色CRUD
		roleGroup := webuiApiGroup.Group("/roles")
		{
			roleGroup.GET("/list", roleController.GetListByID)
			roleGroup.GET("/detail/:id", roleController.GetByID)
			roleGroup.POST("/create", roleController.Create)
			roleGroup.POST("/update", roleController.Update)
			roleGroup.POST("/delete", roleController.Delete)
		}
	}

	// 提供前端静态文件服务
	setupStaticFiles(r)
}

// setupStaticFiles 配置前端静态文件服务
func setupStaticFiles(r *gin.Engine) {
	// 检查前端构建目录是否存在
	webuiDist := "./webui/dist"
	if _, err := os.Stat(webuiDist); os.IsNotExist(err) {
		// 前端目录不存在，跳过静态文件服务
		return
	}

	// 提供静态文件服务
	r.Static("/assets", filepath.Join(webuiDist, "assets"))
	r.StaticFile("/favicon.ico", filepath.Join(webuiDist, "favicon.ico"))

	// 所有非API请求都返回 index.html（支持前端路由）
	r.NoRoute(func(c *gin.Context) {
		// API 请求不处理
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// 返回前端入口文件
		indexPath := filepath.Join(webuiDist, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		}
	})
}
