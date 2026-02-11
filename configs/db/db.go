package db

import (
	"database/sql"
	"fmt"
	"log"
	"sitchi/configs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
func Migrate(dbConfig *configs.DbConfig) error {
	m, err := migrate.New(
		"file://configs/db/migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.SSLMode),
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("数据库已是最新版本")
			return nil
		}
		return err
	}
	log.Println("数据库迁移成功")
	return nil
}
func DownMigrate(dbConfig *configs.DbConfig) error {
	m, err := migrate.New(
		"file://configs/db/migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.SSLMode),
	)
	if err != nil {
		return err
	}
	if err := m.Down(); err != nil {
		return err
	}
	log.Println("数据库回滚成功")
	return nil
}
