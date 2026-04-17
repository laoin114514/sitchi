package resource

import (
	"database/sql"
	"sitchi/internal/dao"
)

// Repository 资源数据访问层
type Repository struct {
	aclDAO *dao.ACLDAO
}

// NewRepository 创建 Repository 实例
func NewRepository(aclDAO *dao.ACLDAO) *Repository {
	if aclDAO == nil {
		aclDAO = dao.NewACLDAO()
	}
	return &Repository{aclDAO: aclDAO}
}

// Resource 资源模型
type Resource struct {
	ID          int64  `json:"id"`
	ModuleID    int64  `json:"module_id"`
	ResCode     string `json:"res_code"`
	ResName     string `json:"res_name"`
	ResType     string `json:"res_type"`
	ParentID    *int64 `json:"parent_id"`
	Path        string `json:"path"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// Insert 创建资源
func (r *Repository) Insert(tx *sql.Tx, moduleID int64, resCode, resName, resType string, parentID *int64, path, description string) (int64, error) {
	var resourceID int64
	err := tx.QueryRow(`
		INSERT INTO resources (module_id, res_code, res_name, res_type, parent_id, path, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, moduleID, resCode, resName, resType, parentID, path, description).Scan(&resourceID)
	if err != nil {
		return 0, err
	}
	return resourceID, nil
}

// Update 更新资源
func (r *Repository) Update(tx *sql.Tx, resourceID, moduleID int64, resName, resType string, parentID *int64, path, description string) error {
	_, err := tx.Exec(`
		UPDATE resources 
		SET res_name = $1, res_type = $2, parent_id = $3, path = $4, description = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6 AND module_id = $7 AND is_deleted = FALSE
	`, resName, resType, parentID, path, description, resourceID, moduleID)
	return err
}

// Delete 软删除资源
func (r *Repository) Delete(tx *sql.Tx, resourceID, moduleID int64) error {
	_, err := tx.Exec(`
		UPDATE resources 
		SET is_deleted = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND module_id = $2
	`, resourceID, moduleID)
	return err
}

// GetByID 根据ID获取资源
func (r *Repository) GetByID(tx *sql.Tx, resourceID, moduleID int64) (*Resource, error) {
	var res Resource
	var parentID sql.NullInt64
	err := tx.QueryRow(`
		SELECT id, module_id, res_code, res_name, res_type, parent_id, path, description, is_active
		FROM resources
		WHERE id = $1 AND module_id = $2 AND is_deleted = FALSE
	`, resourceID, moduleID).Scan(
		&res.ID, &res.ModuleID, &res.ResCode, &res.ResName, &res.ResType,
		&parentID, &res.Path, &res.Description, &res.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if parentID.Valid {
		res.ParentID = &parentID.Int64
	}
	return &res, nil
}

// List 获取资源列表
func (r *Repository) List(tx *sql.Tx, moduleID int64, resType string, parentID *int64) ([]*Resource, error) {
	query := `
		SELECT id, module_id, res_code, res_name, res_type, parent_id, path, description, is_active
		FROM resources
		WHERE module_id = $1 AND is_deleted = FALSE
	`
	args := []interface{}{moduleID}
	argIndex := 2

	if resType != "" {
		query += " AND res_type = $" + string(rune('0'+argIndex))
		args = append(args, resType)
		argIndex++
	}

	if parentID != nil {
		query += " AND parent_id = $" + string(rune('0'+argIndex))
		args = append(args, *parentID)
	} else {
		query += " AND parent_id IS NULL"
	}

	query += " ORDER BY created_at DESC"

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*Resource
	for rows.Next() {
		var res Resource
		var pid sql.NullInt64
		err := rows.Scan(
			&res.ID, &res.ModuleID, &res.ResCode, &res.ResName, &res.ResType,
			&pid, &res.Path, &res.Description, &res.IsActive,
		)
		if err != nil {
			return nil, err
		}
		if pid.Valid {
			res.ParentID = &pid.Int64
		}
		resources = append(resources, &res)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

// CheckCodeExists 检查资源编码是否已存在
func (r *Repository) CheckCodeExists(tx *sql.Tx, moduleID int64, resCode string) (bool, error) {
	var count int
	err := tx.QueryRow(`
		SELECT COUNT(*) FROM resources 
		WHERE module_id = $1 AND res_code = $2 AND is_deleted = FALSE
	`, moduleID, resCode).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
