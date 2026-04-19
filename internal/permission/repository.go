package perm

import (
	"database/sql"
	"sitchi/internal/dao"
)

type Repository struct {
	aclDAO *dao.ACLDAO
}

type Permission struct {
	ID          int64  `json:"id"`
	ModuleID    int64  `json:"module_id"`
	PermCode    string `json:"perm_code"`
	PermName    string `json:"perm_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
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

func (r *Repository) GetPermByID(tx *sql.Tx, moduleID, permID int64) (*Permission, error) {
	var p Permission
	err := tx.QueryRow(`
		SELECT id, module_id, perm_code, perm_name, COALESCE(description,''), is_active, created_at, updated_at
		FROM permissions
		WHERE module_id = $1 AND id = $2 AND is_deleted = FALSE
	`, moduleID, permID).Scan(
		&p.ID, &p.ModuleID, &p.PermCode, &p.PermName, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPermNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetPermList(tx *sql.Tx, moduleID int64) ([]*Permission, error) {
	rows, err := tx.Query(`
		SELECT id, module_id, perm_code, perm_name, COALESCE(description,''), is_active, created_at, updated_at
		FROM permissions
		WHERE module_id = $1 AND is_deleted = FALSE
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []*Permission
	for rows.Next() {
		var p Permission
		err := rows.Scan(
			&p.ID, &p.ModuleID, &p.PermCode, &p.PermName, &p.Description, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		perms = append(perms, &p)
	}
	return perms, rows.Err()
}

func (r *Repository) InsertPerm(tx *sql.Tx, moduleID int64, permCode, permName, description string) (int64, error) {
	var permID int64
	err := tx.QueryRow(`
		INSERT INTO permissions (module_id, perm_code, perm_name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, moduleID, permCode, permName, description).Scan(&permID)
	return permID, err
}

func (r *Repository) UpdatePerm(tx *sql.Tx, permID, moduleID int64, permCode, permName, description string) error {
	_, err := tx.Exec(`
		UPDATE permissions
		SET perm_code = $1, perm_name = $2, description = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND module_id = $5 AND is_deleted=FALSE
	`, permCode, permName, description, permID, moduleID)
	return err
}

func (r *Repository) DeletePerm(tx *sql.Tx, permID, moduleID int64) error {
	_, err := tx.Exec(`
		UPDATE permissions
		SET is_deleted = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND module_id = $2 AND is_deleted=FALSE
	`, permID, moduleID)
	return err
}
