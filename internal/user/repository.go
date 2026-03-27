package user

import (
	"database/sql"
	"fmt"
)

func getModuleIDByCode(tx *sql.Tx, moduleCode string) (int64, error) {
	var moduleID int64
	err := tx.QueryRow(`
		SELECT id
		FROM modules
		WHERE module_code = $1
		  AND is_deleted = FALSE
	`, moduleCode).Scan(&moduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrModuleNotFound
		}
		return 0, err
	}
	return moduleID, nil
}

func insertUser(tx *sql.Tx, moduleID int64, userCode, userName, description, password, email string) (int64, error) {
	var userID int64
	err := tx.QueryRow(`
		INSERT INTO users (user_code, user_name, description, module_id, password, email)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, userCode, userName, description, moduleID, password, email).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func insertModuleUser(tx *sql.Tx, moduleID, userID int64, description string) error {
	_, err := tx.Exec(`
		INSERT INTO module_users (module_id, user_id, description)
		VALUES ($1, $2, $3)
	`, moduleID, userID, description)
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
		if err == sql.ErrNoRows {
			return 0, ErrDefaultRoleNotFound
		}
		return 0, err
	}
	return roleID, nil
}

func bindUserRole(tx *sql.Tx, userID, moduleID, roleID int64) error {
	_, err := tx.Exec(`
		INSERT INTO user_roles (user_id, module_id, role_id)
		VALUES ($1, $2, $3)
	`, userID, moduleID, roleID)
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

func moduleRoleCode(moduleCode, roleCode string) string {
	return fmt.Sprintf("%s:%s", moduleCode, roleCode)
}

type UserAuthRecord struct {
	UserID     int64
	ModuleCode string
	UserCode   string
	Password   string
	Roles      []string
}

func getUserAuthRecordByCode(tx *sql.Tx, moduleCode, userCode string) (*UserAuthRecord, error) {
	var rec UserAuthRecord
	err := tx.QueryRow(`
		SELECT u.id, m.module_code, u.user_code, u.password
		FROM users u
		JOIN modules m ON m.id = u.module_id
		WHERE m.module_code = $1
		  AND u.user_code = $2
		  AND u.is_deleted = FALSE
		  AND u.is_active = TRUE
		  AND m.is_deleted = FALSE
		  AND m.is_active = TRUE
	`, moduleCode, userCode).Scan(&rec.UserID, &rec.ModuleCode, &rec.UserCode, &rec.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	rows, err := tx.Query(`
		SELECT r.role_code
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id AND r.module_id = ur.module_id
		WHERE ur.module_id = (
			SELECT id FROM modules WHERE module_code = $1 AND is_deleted = FALSE
		)
		  AND ur.user_id = $2
		  AND r.is_deleted = FALSE
	`, moduleCode, rec.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rec.Roles = roles
	return &rec, nil
}
