package module

import (
	"database/sql"
	"fmt"
)

func ensureOwnerExists(tx *sql.Tx, userID int64) error {
	var exists bool
	err := tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM users WHERE id = $1 AND is_deleted = FALSE
		)
	`, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrOwnerNotFound
	}
	return nil
}

func insertModule(tx *sql.Tx, ownerUserID int64, moduleCode, moduleName, description string) (int64, error) {
	var owner any
	if ownerUserID > 0 {
		owner = ownerUserID
	} else {
		owner = nil
	}

	var moduleID int64
	err := tx.QueryRow(`
		INSERT INTO modules (module_code, module_name, description, owner_user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, moduleCode, moduleName, description, owner).Scan(&moduleID)
	if err != nil {
		return 0, err
	}
	return moduleID, nil
}

func insertModuleUser(tx *sql.Tx, moduleID, userID int64) error {
	_, err := tx.Exec(`
		INSERT INTO module_users (module_id, user_id, description)
		SELECT $1, $2, $3
		WHERE NOT EXISTS (
			SELECT 1 FROM module_users
			WHERE module_id = $1
			  AND user_id = $2
			  AND is_deleted = FALSE
		)
	`, moduleID, userID, "模块创建者")
	return err
}

func insertDefaultRoles(tx *sql.Tx, moduleID int64, moduleCode string) (map[string]int64, error) {
	roles := []struct {
		Code string
		Desc string
	}{
		{Code: moduleRoleCode(moduleCode, "admin"), Desc: "模块管理员角色"},
		{Code: moduleRoleCode(moduleCode, "normal"), Desc: "模块普通角色"},
	}

	roleIDs := make(map[string]int64, len(roles))
	for _, r := range roles {
		var roleID int64
		err := tx.QueryRow(`
			INSERT INTO roles (module_id, role_code, description)
			VALUES ($1, $2, $3)
			RETURNING id
		`, moduleID, r.Code, r.Desc).Scan(&roleID)
		if err != nil {
			return nil, err
		}
		roleIDs[r.Code] = roleID
	}
	return roleIDs, nil
}

func insertDefaultPermissions(tx *sql.Tx, moduleID int64) (map[string]int64, error) {
	perms := []struct {
		Code string
		Name string
		Desc string
	}{
		{Code: "view", Name: "查看", Desc: "查看资源"},
		{Code: "plus", Name: "新增", Desc: "新增资源数据"},
		{Code: "change", Name: "修改", Desc: "修改资源数据"},
		{Code: "delete", Name: "删除", Desc: "删除资源数据"},
		{Code: "bind", Name: "绑定", Desc: "建立关系绑定"},
		{Code: "unbind", Name: "解绑", Desc: "解除关系绑定"},
	}

	permIDs := make(map[string]int64, len(perms))
	for _, p := range perms {
		var permID int64
		err := tx.QueryRow(`
			INSERT INTO permissions (module_id, perm_code, perm_name, description)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, moduleID, p.Code, p.Name, p.Desc).Scan(&permID)
		if err != nil {
			return nil, err
		}
		permIDs[p.Code] = permID
	}
	return permIDs, nil
}

func insertDefaultResources(tx *sql.Tx, moduleID int64, moduleCode string) (map[string]int64, error) {
	rootCode := fmt.Sprintf("%s:module", moduleCode)
	rootName := fmt.Sprintf("%s 模块", moduleCode)

	var rootID int64
	err := tx.QueryRow(`
		INSERT INTO resources (module_id, res_code, res_name, res_type, parent_id, path, description)
		VALUES ($1, $2, $3, 'module', NULL, $4, $5)
		RETURNING id
	`, moduleID, rootCode, rootName, fmt.Sprintf("/%s", moduleCode), "模块根资源").Scan(&rootID)
	if err != nil {
		return nil, err
	}

	children := []struct {
		Code string
		Name string
		Path string
		Desc string
	}{
		{Code: "modules", Name: "模块表", Path: "/modules", Desc: "模块信息"},
		{Code: "users", Name: "用户表", Path: "/users", Desc: "用户信息"},
		{Code: "roles", Name: "角色表", Path: "/roles", Desc: "角色信息"},
		{Code: "permissions", Name: "权限表", Path: "/permissions", Desc: "权限信息"},
		{Code: "resources", Name: "资源表", Path: "/resources", Desc: "资源信息"},
		{Code: "user_roles", Name: "用户角色关联表", Path: "/user-roles", Desc: "用户角色关系"},
		{Code: "role_permission_resources", Name: "角色权限资源关联表", Path: "/role-permission-resources", Desc: "角色权限资源关系"},
	}

	resIDs := map[string]int64{
		"module_root": rootID,
	}

	for _, c := range children {
		resCode := fmt.Sprintf("%s:%s", moduleCode, c.Code)
		var resID int64
		err = tx.QueryRow(`
			INSERT INTO resources (module_id, res_code, res_name, res_type, parent_id, path, description)
			VALUES ($1, $2, $3, 'table', $4, $5, $6)
			RETURNING id
		`, moduleID, resCode, c.Name, rootID, fmt.Sprintf("/%s%s", moduleCode, c.Path), c.Desc).Scan(&resID)
		if err != nil {
			return nil, err
		}
		resIDs[c.Code] = resID
	}

	return resIDs, nil
}

func bindUserRole(tx *sql.Tx, userID, moduleID, roleID int64) error {
	_, err := tx.Exec(`
		INSERT INTO user_roles (user_id, module_id, role_id)
		VALUES ($1, $2, $3)
	`, userID, moduleID, roleID)
	return err
}

func grantDefaultRolePermissions(tx *sql.Tx, moduleID int64, roleIDs, permIDs, resIDs map[string]int64, moduleCode string) error {
	baseResources := []string{"modules", "users", "roles", "permissions", "resources"}
	relationResources := []string{"user_roles", "role_permission_resources"}

	adminRoleID := roleIDs[moduleRoleCode(moduleCode, "admin")]
	normalRoleID := roleIDs[moduleRoleCode(moduleCode, "normal")]

	if err := grantOne(tx, moduleID, adminRoleID, permIDs["view"], resIDs["module_root"]); err != nil {
		return err
	}

	for _, rc := range baseResources {
		for _, pc := range []string{"view", "plus", "change", "delete"} {
			if err := grantOne(tx, moduleID, adminRoleID, permIDs[pc], resIDs[rc]); err != nil {
				return err
			}
		}
	}

	for _, rc := range relationResources {
		for _, pc := range []string{"view", "bind", "unbind"} {
			if err := grantOne(tx, moduleID, adminRoleID, permIDs[pc], resIDs[rc]); err != nil {
				return err
			}
		}
	}

	if err := grantOne(tx, moduleID, normalRoleID, permIDs["view"], resIDs["module_root"]); err != nil {
		return err
	}

	for _, rc := range append(baseResources, relationResources...) {
		if err := grantOne(tx, moduleID, normalRoleID, permIDs["view"], resIDs[rc]); err != nil {
			return err
		}
	}

	return nil
}

func grantOne(tx *sql.Tx, moduleID, roleID, permID, resID int64) error {
	_, err := tx.Exec(`
		INSERT INTO role_permission_resources (role_id, perm_id, res_id, module_id)
		VALUES ($1, $2, $3, $4)
	`, roleID, permID, resID, moduleID)
	return err
}

func rebuildUserPermResByUser(tx *sql.Tx, moduleID, userID int64) error {
	_, err := tx.Exec(`
		DELETE FROM user_perm_res
		WHERE module_id = $1 AND user_id = $2
	`, moduleID, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO user_perm_res (module_id, user_id, perm_id, res_id)
		SELECT DISTINCT ur.module_id, ur.user_id, rpr.perm_id, rpr.res_id
		FROM user_roles ur
		JOIN role_permission_resources rpr
		  ON rpr.module_id = ur.module_id
		 AND rpr.role_id = ur.role_id
		WHERE ur.module_id = $1
		  AND ur.user_id = $2
	`, moduleID, userID)
	return err
}

func getModuleIDByCode(tx *sql.Tx, moduleCode string) (int64, error) {
	var moduleID int64
	err := tx.QueryRow(`
		SELECT id
		FROM modules
		WHERE module_code = $1
		  AND is_deleted = FALSE
	`, moduleCode).Scan(&moduleID)
	if err != nil {
		return 0, err
	}
	return moduleID, nil
}

func updateModuleOwner(tx *sql.Tx, moduleID, ownerUserID int64) error {
	_, err := tx.Exec(`
		UPDATE modules
		SET owner_user_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, ownerUserID, moduleID)
	return err
}

func getRoleIDByCode(tx *sql.Tx, moduleID int64, roleCode string) (int64, error) {
	var roleID int64
	err := tx.QueryRow(`
		SELECT id
		FROM roles
		WHERE module_id = $1
		  AND role_code = $2
		  AND is_deleted = FALSE
	`, moduleID, roleCode).Scan(&roleID)
	if err != nil {
		return 0, err
	}
	return roleID, nil
}

func moduleRoleCode(moduleCode, roleCode string) string {
	return fmt.Sprintf("%s:%s", moduleCode, roleCode)
}
