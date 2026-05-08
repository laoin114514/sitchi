package module

import (
	"database/sql"
	"fmt"

	"sitchi/internal/dao"
)

type Module struct {
	ID          int64  `json:"id"`
	ModuleCode  string `json:"module_code"`
	ModuleName  string `json:"module_name"`
	Description string `json:"description"`
	OwnerUserID int64  `json:"owner_user_id"`
	IsActive    bool   `json:"is_active"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type Repository struct {
	aclDAO *dao.ACLDAO
}

func NewRepository(aclDAO *dao.ACLDAO) *Repository {
	if aclDAO == nil {
		aclDAO = dao.NewACLDAO()
	}
	return &Repository{aclDAO: aclDAO}
}

func (r *Repository) EnsureOwnerExists(tx *sql.Tx, userID int64) error {
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

func (r *Repository) InsertModule(tx *sql.Tx, ownerUserID int64, moduleCode, moduleName, description string) (int64, error) {
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

func (r *Repository) InsertModuleUser(tx *sql.Tx, moduleID, userID int64) error {
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

func (r *Repository) InsertDefaultRoles(tx *sql.Tx, moduleID int64, moduleCode string) (map[string]int64, error) {
	roles := []struct {
		Code string
		Desc string
	}{
		{Code: r.ModuleRoleCode(moduleCode, "admin"), Desc: "模块管理员角色"},
		{Code: r.ModuleRoleCode(moduleCode, "normal"), Desc: "模块普通角色"},
	}

	roleIDs := make(map[string]int64, len(roles))
	for _, item := range roles {
		var roleID int64
		err := tx.QueryRow(`
			INSERT INTO roles (module_id, role_code, description)
			VALUES ($1, $2, $3)
			RETURNING id
		`, moduleID, item.Code, item.Desc).Scan(&roleID)
		if err != nil {
			return nil, err
		}
		roleIDs[item.Code] = roleID
	}
	return roleIDs, nil
}

func (r *Repository) InsertDefaultPermissions(tx *sql.Tx, moduleID int64) (map[string]int64, error) {
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
	for _, item := range perms {
		var permID int64
		err := tx.QueryRow(`
			INSERT INTO permissions (module_id, perm_code, perm_name, description)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, moduleID, item.Code, item.Name, item.Desc).Scan(&permID)
		if err != nil {
			return nil, err
		}
		permIDs[item.Code] = permID
	}
	return permIDs, nil
}

func (r *Repository) InsertDefaultResources(tx *sql.Tx, moduleID int64, moduleCode string) (map[string]int64, error) {
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

	resIDs := map[string]int64{"module_root": rootID}
	for _, item := range children {
		resCode := fmt.Sprintf("%s:%s", moduleCode, item.Code)
		var resID int64
		err = tx.QueryRow(`
			INSERT INTO resources (module_id, res_code, res_name, res_type, parent_id, path, description)
			VALUES ($1, $2, $3, 'table', $4, $5, $6)
			RETURNING id
		`, moduleID, resCode, item.Name, rootID, fmt.Sprintf("/%s%s", moduleCode, item.Path), item.Desc).Scan(&resID)
		if err != nil {
			return nil, err
		}
		resIDs[item.Code] = resID
	}

	return resIDs, nil
}

func (r *Repository) BindUserRole(tx *sql.Tx, userID, moduleID, roleID int64) error {
	return r.aclDAO.BindUserRole(tx, userID, moduleID, roleID)
}

func (r *Repository) GrantDefaultRolePermissions(tx *sql.Tx, moduleID int64, roleIDs, permIDs, resIDs map[string]int64, moduleCode string) error {
	baseResources := []string{"modules", "users", "roles", "permissions", "resources"}
	relationResources := []string{"user_roles", "role_permission_resources"}

	adminRoleID := roleIDs[r.ModuleRoleCode(moduleCode, "admin")]
	normalRoleID := roleIDs[r.ModuleRoleCode(moduleCode, "normal")]

	if err := r.grantOne(tx, moduleID, adminRoleID, permIDs["view"], resIDs["module_root"]); err != nil {
		return err
	}

	for _, rc := range baseResources {
		for _, pc := range []string{"view", "plus", "change", "delete"} {
			if err := r.grantOne(tx, moduleID, adminRoleID, permIDs[pc], resIDs[rc]); err != nil {
				return err
			}
		}
	}

	for _, rc := range relationResources {
		for _, pc := range []string{"view", "bind", "unbind"} {
			if err := r.grantOne(tx, moduleID, adminRoleID, permIDs[pc], resIDs[rc]); err != nil {
				return err
			}
		}
	}

	if err := r.grantOne(tx, moduleID, normalRoleID, permIDs["view"], resIDs["module_root"]); err != nil {
		return err
	}

	for _, rc := range append(baseResources, relationResources...) {
		if err := r.grantOne(tx, moduleID, normalRoleID, permIDs["view"], resIDs[rc]); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) grantOne(tx *sql.Tx, moduleID, roleID, permID, resID int64) error {
	_, err := tx.Exec(`
		INSERT INTO role_permission_resources (role_id, perm_id, res_id, module_id)
		VALUES ($1, $2, $3, $4)
	`, roleID, permID, resID, moduleID)
	return err
}

func (r *Repository) RebuildUserPermResByUser(tx *sql.Tx, moduleID, userID int64) error {
	return r.aclDAO.RebuildUserPermResByUser(tx, moduleID, userID)
}

func (r *Repository) GetModuleIDByCode(tx *sql.Tx, moduleCode string) (int64, error) {
	return r.aclDAO.GetModuleIDByCode(tx, moduleCode)
}

func (r *Repository) UpdateModuleOwner(tx *sql.Tx, moduleID, ownerUserID int64) error {
	_, err := tx.Exec(`
		UPDATE modules
		SET owner_user_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, ownerUserID, moduleID)
	return err
}

func (r *Repository) GetRoleIDByCode(tx *sql.Tx, moduleID int64, roleCode string) (int64, error) {
	return r.aclDAO.GetRoleIDByCode(tx, moduleID, roleCode)
}

func (r *Repository) ModuleRoleCode(moduleCode, roleCode string) string {
	return r.aclDAO.ModuleRoleCode(moduleCode, roleCode)
}

func (r *Repository) EnsureSuperRole(tx *sql.Tx, moduleID int64, moduleCode string) (int64, error) {
	superRoleCode := r.ModuleRoleCode(moduleCode, "super")

	var roleID int64
	err := tx.QueryRow(`
		SELECT id
		FROM roles
		WHERE module_id = $1
		  AND role_code = $2
		  AND is_deleted = FALSE
	`, moduleID, superRoleCode).Scan(&roleID)
	if err == nil {
		return roleID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	err = tx.QueryRow(`
		INSERT INTO roles (module_id, role_code, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, moduleID, superRoleCode, "超级管理员角色").Scan(&roleID)
	if err != nil {
		return 0, err
	}

	return roleID, nil
}

func (r *Repository) GetModuleByID(tx *sql.Tx, moduleID int64) (*Module, error) {
	var m Module
	var owner any
	err := tx.QueryRow(`
		SELECT id, module_code, module_name, COALESCE(description,''), COALESCE(owner_user_id,0), is_active, updated_at, created_at
		FROM modules
		WHERE id = $1 AND is_deleted = FALSE
	`, moduleID).Scan(&m.ID, &m.ModuleCode, &m.ModuleName, &m.Description, &owner, &m.IsActive, &m.UpdatedAt, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrModuleNotFound
		}
		return nil, err
	}
	m.OwnerUserID = owner.(int64)
	return &m, nil
}

func (r *Repository) GetModuleList(tx *sql.Tx) ([]*Module, error) {
	rows, err := tx.Query(`
		SELECT id, module_code, module_name, COALESCE(description,''), COALESCE(owner_user_id,0), is_active, updated_at, created_at
		FROM modules
		WHERE is_deleted = FALSE
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []*Module
	for rows.Next() {
		var m Module
		var owner any
		err := rows.Scan(&m.ID, &m.ModuleCode, &m.ModuleName, &m.Description, &owner, &m.IsActive, &m.UpdatedAt, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		m.OwnerUserID = owner.(int64)
		modules = append(modules, &m)
	}
	return modules, rows.Err()
}

func (r *Repository) UpdateModule(tx *sql.Tx, moduleID int64, moduleName, description string) error {
	_, err := tx.Exec(`
		UPDATE modules
		SET module_name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND is_deleted = FALSE
	`, moduleName, description, moduleID)
	return err
}

func (r *Repository) DeleteModule(tx *sql.Tx, moduleID int64) error {
	_, err := tx.Exec(`
		UPDATE modules
		SET is_deleted = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_deleted = FALSE
	`, moduleID)
	return err
}
