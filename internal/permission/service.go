package perm

import (
	"database/sql"
	"errors"
	"sitchi/internal/dao"
	"strings"
)

type CreatePermParams struct {
	ModuleCode  string
	PermCode    string
	PermName    string
	Description string
}

type UpdatePermParams struct {
	ModuleCode  string
	PermID      int64
	PermCode    string
	PermName    string
	Description string
}

var (
	ErrPermNotFound     = errors.New("permission not found")
	ErrModuleNotFound   = errors.New("module not found")
	ErrInvalidParams    = errors.New("invalid permission params")
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

func (s *Service) CreatePerm(params CreatePermParams) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	permCode := strings.TrimSpace(params.PermCode)
	permName := strings.TrimSpace(params.PermName)

	if moduleCode == "" || permCode == "" || permName == "" {
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

	permID, err := s.repo.InsertPerm(tx, moduleID, permCode, permName, strings.TrimSpace(params.Description))
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return permID, nil
}

func (s *Service) GetByID(moduleCode string, permID int64) (*Permission, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || permID <= 0 {
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

	perm, err := s.repo.GetPermByID(tx, moduleID, permID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return perm, nil
}

func (s *Service) GetListByID(moduleCode string) ([]*Permission, error) {
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

	permList, err := s.repo.GetPermListByID(tx, moduleID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return permList, nil
}

func (s *Service) Update(params UpdatePermParams) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	moduleCode := strings.TrimSpace(params.ModuleCode)
	permCode := strings.TrimSpace(params.PermCode)
	permName := strings.TrimSpace(params.PermName)
	if moduleCode == "" || permCode == "" || permName == "" || params.PermID <= 0 {
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

	if err = s.repo.UpdatePerm(tx, params.PermID, moduleID, permCode, permName, strings.TrimSpace(params.Description)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(moduleCode string, permID int64) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || permID <= 0 {
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

	if err = s.repo.DeletePerm(tx, permID, moduleID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
