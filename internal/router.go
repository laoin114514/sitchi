package internal

import (
	"net/http"
	"os"
	"path/filepath"
	perm "sitchi/internal/permission"
	"sitchi/internal/role"
	"sitchi/internal/user"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	userController := user.NewController()
	roleController := role.NewController()
	permController := perm.NewController()

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
		webuiApiGroup.POST("/roles", roleController.Create)
		webuiApiGroup.GET("/roles", roleController.GetListByID)
		webuiApiGroup.GET("/roles/:id", roleController.GetByID)
		webuiApiGroup.POST("/roles/update/:id", roleController.Update)
		webuiApiGroup.POST("/roles/delete/:id", roleController.Delete)

		// 权限CRUD
		webuiApiGroup.POST("/permissions", permController.Create)
		webuiApiGroup.GET("/permissions", permController.GetListByID)
		webuiApiGroup.GET("/permissions/:id", permController.GetByID)
		webuiApiGroup.POST("/permissions/update/:id", permController.Update)
		webuiApiGroup.POST("/permissions/delete/:id", permController.Delete)
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
