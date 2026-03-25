package main

import (
	"log"
	"os"
	"sitchi/configs"
	"sitchi/configs/db"
	"sitchi/internal/module"
	"sitchi/internal/user"
	"strconv"
)

func init() {
	//获取配置文件路径
	path := configs.CheckMode()

	//加载配置文件
	err := configs.LoadConfig(path)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	err = db.InitDb(&configs.AppConfig.Db)
	if err != nil {
		log.Fatal(err.Error())
	}
}
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("请输入命令")
	}
	switch os.Args[1] {
	case "migrate":
		migrate()
		return
	case "downMigrate":
		downMigrate()
		return
	case "runServer":
		runServer()
		return
	case "initSystem":
		migrate()
		initSystem()
		return
	case "migrateStep":
		stepCount, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("参数错误: %v", err)
		}
		migrateStep(stepCount)
		return
	}

	log.Fatalf("请输入指令: migrate/downMigrate/runServer/initSystem/initModule/migrateStep")
}
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
func runServer() {

	//启动服务写在这

}
func initSystem() {
	adminPassword := "admin123456"
	if len(os.Args) >= 3 {
		adminPassword = os.Args[2]
	}

	moduleID, err := module.CreateModule(module.CreateModuleParams{
		UserID:            0,
		ModuleCode:        "admin",
		ModuleName:        "管理模块",
		ModuleDescription: "系统管理与鉴权模块",
	})
	if err != nil {
		log.Fatalf("初始化 admin 模块失败: %v", err)
	}

	adminUserID, err := user.CreateUser(user.CreateUserParams{
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

	if err = module.SetModuleOwnerAndGrantAdmin("admin", adminUserID); err != nil {
		log.Fatalf("设置 admin 模块 owner 与管理员角色失败: %v", err)
	}

	log.Printf("系统初始化成功：admin_module_id=%d, admin_user_id=%d, admin_password=%s", moduleID, adminUserID, adminPassword)
}
