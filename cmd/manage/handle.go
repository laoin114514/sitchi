package main

import (
	"fmt"
	"log"
	"os"
	"sitchi/configs"
	"sitchi/configs/db"
	"sitchi/internal"
	"sitchi/internal/dao"
	"sitchi/internal/module"
	"sitchi/internal/user"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func migrate() {
	err := db.Migrate(&configs.AppConfig.Db)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
}
func downMigrate() {
	err := db.DownMigrate(&configs.AppConfig.Db)
	if err != nil {
		log.Fatalf("数据库回滚失败: %v", err)
	}
}
func migrateStep(step int) {
	err := db.MigrateStep(&configs.AppConfig.Db, step)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
}
func initSystem() {
	adminPassword := "admin123456"
	if len(os.Args) >= 3 {
		adminPassword = os.Args[2]
	}

	aclDAO := dao.NewACLDAO()
	moduleService := module.NewService(db.Pool, module.NewRepository(aclDAO))
	userService := user.NewService(db.Pool, user.NewRepository(aclDAO))

	moduleID, err := moduleService.CreateModule(module.CreateModuleParams{
		UserID:            0,
		ModuleCode:        "admin",
		ModuleName:        "管理模块",
		ModuleDescription: "系统管理与鉴权模块",
	})
	if err != nil {
		log.Fatalf("初始化 admin 模块失败: %v", err)
	}

	adminUserID, err := userService.CreateUser(user.CreateUserParams{
		ModuleCode:  "admin",
		UserCode:    "admin",
		UserName:    "管理员",
		Password:    adminPassword,
		Email:       "admin@sitchi.local",
		Description: "系统初始化管理员账号",
	})
	if err != nil {
		log.Fatalf("创建 admin 账号失败: %v", err)
	}

	if err = moduleService.SetModuleOwnerAndGrantAdmin("admin", adminUserID); err != nil {
		log.Fatalf("设置 admin 模块 owner 与管理员角色失败: %v", err)
	}

	log.Printf("系统初始化成功：admin_module_id=%d, admin_user_id=%d, admin_password=%s", moduleID, adminUserID, adminPassword)
}

func runServer() {
	gin.SetMode(gin.ReleaseMode)
	if configs.AppConfig != nil && configs.AppConfig.Dev {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     configs.AppConfig.Server.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	internal.InitRouter(r)

	addr := fmt.Sprintf("%s:%d", configs.AppConfig.Server.Host, configs.AppConfig.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
