package module

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NullInt64 自定义类型，用于处理JSON序列化时的null值
type NullInt64 sql.NullInt64

// MarshalJSON 实现json.Marshaler接口
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Int64)
}

// Scan 实现sql.Scanner接口
func (n *NullInt64) Scan(value interface{}) error {
	var nullInt sql.NullInt64
	if err := nullInt.Scan(value); err != nil {
		return err
	}
	*n = NullInt64(nullInt)
	return nil
}

// Module 模块信息结构
type Module struct {
	ID          int64     `json:"id"`
	ModuleCode  string    `json:"module_code"`
	ModuleName  string    `json:"module_name"`
	Description string    `json:"description"`
	OwnerUserID NullInt64 `json:"owner_user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsDeleted   bool      `json:"is_deleted"`
}
