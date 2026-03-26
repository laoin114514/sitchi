package module

import (
	"errors"
	"strings"

	"sitchi/configs/db"
)

type CreateModuleParams struct {
	UserID            int64
	ModuleCode        string
	ModuleName        string
	ModuleDescription string
}

var (
	ErrInvalidParams    = errors.New("invalid create module params")
	ErrDBNotInitialized = errors.New("db pool is not initialized")
	ErrOwnerNotFound    = errors.New("owner user not found")
)

// CreateModule 创建模块并初始化 ACL 基础数据：
// 1) modules（owner_user_id = UserID）
// 2) module_users（加入模块成员）
// 3) roles（admin/normal）
// 4) permissions（view/plus/change/delete/bind/unbind）
// 5) resources（<module_code>:module 根资源 + 固定子资源）
// 6) 将 owner 绑定到 admin 角色
// 7) 初始化 admin/normal 默认授权并写入 owner 的 user_perm_res
func CreateModule(params CreateModuleParams) (int64, error) {
	if db.Pool == nil {
		return 0, ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	moduleName := strings.TrimSpace(params.ModuleName)
	if moduleCode == "" || moduleName == "" {
		return 0, ErrInvalidParams
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

	ownerUserID := params.UserID
	if ownerUserID > 0 {
		if err = ensureOwnerExists(tx, ownerUserID); err != nil {
			return 0, err
		}
	}

	moduleID, err := insertModule(tx, ownerUserID, moduleCode, moduleName, strings.TrimSpace(params.ModuleDescription))
	if err != nil {
		return 0, err
	}

	if ownerUserID > 0 {
		if err = insertModuleUser(tx, moduleID, ownerUserID); err != nil {
			return 0, err
		}
	}

	roleIDs, err := insertDefaultRoles(tx, moduleID, moduleCode)
	if err != nil {
		return 0, err
	}

	permIDs, err := insertDefaultPermissions(tx, moduleID)
	if err != nil {
		return 0, err
	}

	resIDs, err := insertDefaultResources(tx, moduleID, moduleCode)
	if err != nil {
		return 0, err
	}

	if ownerUserID > 0 {
		if err = bindUserRole(tx, ownerUserID, moduleID, roleIDs[moduleRoleCode(moduleCode, "admin")]); err != nil {
			return 0, err
		}
	}

	if err = grantDefaultRolePermissions(tx, moduleID, roleIDs, permIDs, resIDs, moduleCode); err != nil {
		return 0, err
	}

	if ownerUserID > 0 {
		if err = rebuildUserPermResByUser(tx, moduleID, ownerUserID); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return moduleID, nil
}

// SetModuleOwnerAndGrantAdmin 设置模块 owner，并将该用户加入模块及绑定模块 admin 角色。
func SetModuleOwnerAndGrantAdmin(moduleCode string, ownerUserID int64) error {
	if db.Pool == nil {
		return ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || ownerUserID <= 0 {
		return ErrInvalidParams
	}

	tx, err := db.Pool.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = ensureOwnerExists(tx, ownerUserID); err != nil {
		return err
	}

	moduleID, err := getModuleIDByCode(tx, moduleCode)
	if err != nil {
		return err
	}

	if err = updateModuleOwner(tx, moduleID, ownerUserID); err != nil {
		return err
	}

	if err = insertModuleUser(tx, moduleID, ownerUserID); err != nil {
		return err
	}

	superRoleID, err := ensureSuperRole(tx, moduleID, moduleCode)
	if err != nil {
		return err
	}

	adminRoleID, err := getRoleIDByCode(tx, moduleID, moduleRoleCode(moduleCode, "admin"))
	if err != nil {
		return err
	}

	if err = bindUserRole(tx, ownerUserID, moduleID, adminRoleID); err != nil {
		return err
	}

	if err = bindUserRole(tx, ownerUserID, moduleID, superRoleID); err != nil {
		return err
	}

	if err = rebuildUserPermResByUser(tx, moduleID, ownerUserID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
