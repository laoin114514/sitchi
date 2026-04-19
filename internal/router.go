package internal

import (
	"net/http"
	"sitchi/internal/role"
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
}
