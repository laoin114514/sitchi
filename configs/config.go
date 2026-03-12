package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Db     DbConfig     `yaml:"db"`
	Auth   AuthConfig   `yaml:"auth"`
	Dev    bool         `yaml:"dev"`
}
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}
type DbConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type AuthConfig struct {
	JwtSecret       string `yaml:"jwt_secret"`
	TokenTTLMinutes int    `yaml:"token_ttl_minutes"`
}

var AppConfig *Config

func LoadConfig(path string) error {
	var config Config
	ymlConfig, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(ymlConfig, &config)
	if err != nil {
		return err
	}
	if config.Dev {
		log.Println("开发模式")
	} else {
		log.Println("生产模式")
	}
	AppConfig = &config
	return nil
}
func CheckMode() string {
	var path string
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	mode := os.Getenv("MODE")
	if mode == "dev" {
		path = "configs/configs.dev.yml"
	} else if mode == "prod" {
		path = "configs/configs.prod.yml"
	} else {
		log.Fatalf("环境变量MODE错误: %v,请检查.env文件", mode)
	}
	return path
}
