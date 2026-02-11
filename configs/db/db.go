package db

import (
	"database/sql"
	"fmt"
	"sitchi/configs"

	_ "github.com/lib/pq"
)

var Pool *sql.DB

func InitDb(dbConfig *configs.DbConfig) error {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Dbname, dbConfig.SSLMode))
	if err != nil {
		return err
	}
	// 验证数据库连接是否可用
	if err := db.Ping(); err != nil {
		return err
	}
	Pool = db
	return nil
}
