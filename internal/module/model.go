package module

import "time"

// Module 模块信息结构
type Module struct {
	ID          int64     `json:"id"`
	ModuleCode  string    `json:"module_code"`
	ModuleName  string    `json:"module_name"`
	Description string    `json:"description"`
	OwnerUserID int64     `json:"owner_user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsDeleted   bool      `json:"is_deleted"`
}
