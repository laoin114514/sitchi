package role

import (
	"database/sql"
	"errors"
	"sitchi/internal/dao"
	"strings"
)

type CreateRoleParams struct {
	ModuleCode  string
	RoleCode    string
	Description string
}

type UpdateRoleParams struct {
	ModuleCode  string
	RoleID      int64
	RoleCode    string
	Description string
}

var (
	ErrRoleNotFound     = errors.New("role not found")
	ErrModuleNotFound   = errors.New("module not found")
	ErrInvalidParams    = errors.New("invalid role params")
	ErrDBNotInitialized = errors.New("db pool is not initialized")
)

type Service struct {
	db   *sql.DB
	repo *Repository
}

func NewService(db *sql.DB, repo *Repository) *Service {
	if repo == nil {
		repo = NewRepository(dao.NewACLDAO())
	}
	return &Service{db: db, repo: repo}
}

func (s *Service) CreateRole(params CreateRoleParams) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	roleCode := strings.TrimSpace(params.RoleCode)

	if moduleCode == "" || roleCode == "" {
		return 0, ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	moduleID, err := s.repo.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return 0, err
	}

	roleID, err := s.repo.InsertRole(tx, moduleID, roleCode, strings.TrimSpace(params.Description))
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return roleID, nil
}

func (s *Service) GetByID(moduleCode string, roleID int64) (*Role, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || roleID <= 0 {
		return nil, ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	moduleID, err := s.repo.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return nil, err
	}

	role, err := s.repo.GetRoleByID(tx, moduleID, roleID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *Service) GetListByID(moduleCode string) ([]*Role, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" {
		return nil, ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	moduleID, err := s.repo.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return nil, err
	}

	roleList, err := s.repo.GetRoleListByID(tx, moduleID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return roleList, nil
}

func (s *Service) Update(params UpdateRoleParams) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	roleCode := strings.TrimSpace(params.RoleCode)
	if moduleCode == "" || roleCode == "" || params.RoleID <= 0 {
		return ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	moduleID, err := s.repo.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return err
	}

	if err = s.repo.UpdateRole(tx, params.RoleID, moduleID, roleCode, strings.TrimSpace(params.Description)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(moduleCode string, roleID int64) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || roleID <= 0 {
		return ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	moduleID, err := s.repo.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return err
	}

	if err = s.repo.DeleteRole(tx, roleID, moduleID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
