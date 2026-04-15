package dao

import (
	"database/sql"
	"fmt"
)

var ACLDAOInstance = NewACLDAO()

type ACLDAO struct{}

func NewACLDAO() *ACLDAO {
	return &ACLDAO{}
}

func (d *ACLDAO) ModuleRoleCode(moduleCode, roleCode string) string {
	return fmt.Sprintf("%s:%s", moduleCode, roleCode)
}

func (d *ACLDAO) GetModuleIDByCode(tx *sql.Tx, moduleCode string) (int64, error) {
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

func (d *ACLDAO) GetRoleIDByCode(tx *sql.Tx, moduleID int64, roleCode string) (int64, error) {
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

func (d *ACLDAO) BindUserRole(tx *sql.Tx, userID, moduleID, roleID int64) error {
	_, err := tx.Exec(`
		INSERT INTO user_roles (user_id, module_id, role_id)
		VALUES ($1, $2, $3)
	`, userID, moduleID, roleID)
	return err
}

func (d *ACLDAO) RebuildUserPermResByUser(tx *sql.Tx, moduleID, userID int64) error {
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
