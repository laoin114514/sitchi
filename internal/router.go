package internal

import (
	"net/http"
	"sitchi/internal/resource"
	"sitchi/internal/user"
	"time"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	userController := user.NewController()
	resourceController := resource.NewController()

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

		// 资源管理接口
		webuiApiGroup.GET("/resources", resourceController.List)
		webuiApiGroup.POST("/resources", resourceController.Create)
		webuiApiGroup.GET("/resources/:id", resourceController.GetByID)
		webuiApiGroup.PUT("/resources/:id", resourceController.Update)
		webuiApiGroup.DELETE("/resources/:id", resourceController.Delete)
	}
}
