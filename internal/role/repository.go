package role

import (
	"database/sql"
	"sitchi/internal/dao"
)

type Repository struct {
	aclDAO *dao.ACLDAO
}

// 对应roles表的字段
type Role struct {
	ID          int64  `json:"id"`
	ModuleID    int64  `json:"module_id"`
	RoleCode    string `json:"role_code"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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

func (r *Repository) GetRoleByID(tx *sql.Tx, moduleID, roleID int64) (*Role, error) {
	var rol Role
	err := tx.QueryRow(`
		SELECT id, module_id, role_code, COALESCE(description,''), is_active, updated_at, created_at
		FROM roles
		WHERE id = $1 AND module_id = $2 AND is_deleted=FALSE
	`, roleID, moduleID).Scan(&rol.ID, &rol.ModuleID, &rol.RoleCode, &rol.Description, &rol.IsActive, &rol.UpdatedAt, &rol.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &rol, nil
}

func (r *Repository) GetRoleListByID(tx *sql.Tx, moduleID int64) ([]*Role, error) {
	rows, err := tx.Query(`
		SELECT id, module_id, role_code, COALESCE(description,''), is_active, updated_at, created_at
		FROM roles
		WHERE module_id = $1 AND is_deleted=FALSE
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*Role
	for rows.Next() {
		var rol Role
		err := rows.Scan(&rol.ID, &rol.ModuleID, &rol.RoleCode, &rol.Description, &rol.IsActive, &rol.UpdatedAt, &rol.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &rol)
	}
	return roles, rows.Err()
}

func (r *Repository) InsertRole(tx *sql.Tx, moduleID int64, roleCode, description string) (int64, error) {
	var roleID int64
	err := tx.QueryRow(`
		INSERT INTO roles (module_id, role_code, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, moduleID, roleCode, description).Scan(&roleID)
	return roleID, err
}

func (r *Repository) UpdateRole(tx *sql.Tx, roleID, moduleID int64, roleCode, description string) error {
	_, err := tx.Exec(`
		UPDATE roles
		SET role_code = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND module_id = $4 AND is_deleted=FALSE
	`, roleCode, description, roleID, moduleID)
	return err
}

func (r *Repository) DeleteRole(tx *sql.Tx, roleID, moduleID int64) error {
	_, err := tx.Exec(`
		UPDATE roles
		SET is_deleted = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND module_id = $2 AND is_deleted=FALSE
	`, roleID, moduleID)
	return err
}
