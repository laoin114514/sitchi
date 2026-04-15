package user

import (
	"database/sql"

	"sitchi/internal/dao"
)

type Repository struct {
	aclDAO *dao.ACLDAO
}

func NewRepository(aclDAO *dao.ACLDAO) *Repository {
	if aclDAO == nil {
		aclDAO = dao.NewACLDAO()
	}
	return &Repository{aclDAO: aclDAO}
}

func (r *Repository) GetModuleIDByCode(tx *sql.Tx, moduleCode string) (int64, error) {
	moduleID, err := r.aclDAO.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrModuleNotFound
		}
		return 0, err
	}
	return moduleID, nil
}

func (r *Repository) InsertUser(tx *sql.Tx, moduleID int64, userCode, userName, description, password, email string) (int64, error) {
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

func (r *Repository) InsertModuleUser(tx *sql.Tx, moduleID, userID int64, description string) error {
	_, err := tx.Exec(`
		INSERT INTO module_users (module_id, user_id, description)
		VALUES ($1, $2, $3)
	`, moduleID, userID, description)
	return err
}

func (r *Repository) GetRoleIDByCode(tx *sql.Tx, moduleID int64, roleCode string) (int64, error) {
	roleID, err := r.aclDAO.GetRoleIDByCode(tx, moduleID, roleCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrDefaultRoleNotFound
		}
		return 0, err
	}
	return roleID, nil
}

func (r *Repository) BindUserRole(tx *sql.Tx, userID, moduleID, roleID int64) error {
	return r.aclDAO.BindUserRole(tx, userID, moduleID, roleID)
}

func (r *Repository) RebuildUserPermResByUser(tx *sql.Tx, moduleID, userID int64) error {
	return r.aclDAO.RebuildUserPermResByUser(tx, moduleID, userID)
}

func (r *Repository) ModuleRoleCode(moduleCode, roleCode string) string {
	return r.aclDAO.ModuleRoleCode(moduleCode, roleCode)
}

type UserAuthRecord struct {
	UserID     int64
	ModuleCode string
	UserCode   string
	Password   string
	Roles      []string
}

func (r *Repository) GetUserAuthRecordByCode(tx *sql.Tx, moduleCode, userCode string) (*UserAuthRecord, error) {
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
