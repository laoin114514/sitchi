package main

import (
	"log"
	"os"
	"sitchi/configs"
	"sitchi/configs/db"
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
	switch os.Args[1] {
	case "migrate":
		migrate()
	case "downMigrate":
		downMigrate()
	case "runServer":
		runServer()
	case "migrateStep":
		stepCount, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("参数错误: %v", err)
		}
		migrateStep(stepCount)
	}
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
