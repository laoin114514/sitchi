package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server 	  ServerConfig 	  `yaml:"server"`
	Db     	  DbConfig     	  `yaml:"db"`
	Auth   	  AuthConfig   	  `yaml:"auth"`
	Dev    	  bool         	  `yaml:"dev"`
	Log    	  LogConfig    	  `yaml:"log"` // 对应配置文件中的 log 节点
	RateLimit RateLimitConfig `yaml:"rate_limit"`
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

// LogConfig 日志核心配置
type LogConfig struct {
	Level           string     `yaml:"level"`             // 日志级别（debug/info/warn/error/fatal）
	TimestampFormat string     `yaml:"timestamp_format"`  // 时间戳格式
	OutputToConsole bool       `yaml:"output_to_console"` // 是否输出到控制台
	OutputToFile    bool       `yaml:"output_to_file"`    // 是否输出到文件
	File            FileConfig `yaml:"file"`              // 文件输出配置
}

// FileConfig 日志文件配置（基于lumberjack的切割规则）
type FileConfig struct {
	Filename   string `yaml:"filename"`    // 日志文件路径（如 ./logs/app.log）
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小（单位：MB）
	MaxBackups int    `yaml:"max_backups"` // 保留的日志备份文件数
	MaxAge     int    `yaml:"max_age"`     // 日志文件保留天数
	Compress   bool   `yaml:"compress"`    // 是否压缩备份文件
}

// 限速器配置 - 不同限速类型
type RateLimitConfig struct {
	InternalApis     RateLimitItem `yaml:"internal_apis"`      // 内部 API 限流配置
	RegularUserApis  RateLimitItem `yaml:"regular_user_apis"`  // 普通用户 API 限流配置
	HighFreqApis     RateLimitItem `yaml:"high_freq_apis"`	   // 高频接口
	OpenApis     	 RateLimitItem `yaml:"open_apis"`		   // 开放 API（给第三方用）
	AntiCrawlerApis  RateLimitItem `yaml:"anti_crawler_apis"`  // 反爬虫
}

type RateLimitItem struct {
	Rate     int `yaml:"rate"`     // 限流速率
	Capacity int `yaml:"capacity"` // 令牌桶容量（或最大突发量）
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
