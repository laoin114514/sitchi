package user

import (
	"errors"
	"strings"

	"sitchi/configs/db"
)

type CreateUserParams struct {
	ModuleCode  string
	UserCode    string
	UserName    string
	Password    string
	Email       string
	Description string
}

var (
	ErrInvalidCreateUserParams = errors.New("invalid create user params")
	ErrDBNotInitialized        = errors.New("db pool is not initialized")
	ErrModuleNotFound          = errors.New("module not found")
	ErrDefaultRoleNotFound     = errors.New("default role not found")
)

// CreateUser 创建用户并绑定模块默认角色（module_code:normal），最后刷新用户权限缓存。
func CreateUser(params CreateUserParams) (int64, error) {
	if db.Pool == nil {
		return 0, ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	userCode := strings.TrimSpace(params.UserCode)
	userName := strings.TrimSpace(params.UserName)
	password := strings.TrimSpace(params.Password)

	if moduleCode == "" || userCode == "" || userName == "" || password == "" {
		return 0, ErrInvalidCreateUserParams
	}

	tx, err := db.Pool.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	moduleID, err := getModuleIDByCode(tx, moduleCode)
	if err != nil {
		return 0, err
	}

	userID, err := insertUser(tx, moduleID, userCode, userName, strings.TrimSpace(params.Description), password, strings.TrimSpace(params.Email))
	if err != nil {
		return 0, err
	}

	if err = insertModuleUser(tx, moduleID, userID, "模块用户"); err != nil {
		return 0, err
	}

	normalRoleCode := moduleRoleCode(moduleCode, "normal")
	roleID, err := getRoleIDByCode(tx, moduleID, normalRoleCode)
	if err != nil {
		return 0, err
	}

	if err = bindUserRole(tx, userID, moduleID, roleID); err != nil {
		return 0, err
	}

	if err = rebuildUserPermResByUser(tx, moduleID, userID); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}
