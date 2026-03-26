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
