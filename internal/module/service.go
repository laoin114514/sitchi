package module

import (
	"database/sql"
	"errors"
	"strings"

	"sitchi/internal/dao"
)

type CreateModuleParams struct {
	UserID            int64
	ModuleCode        string
	ModuleName        string
	ModuleDescription string
}

type UpdateModuleParams struct {
	ModuleID    int64
	ModuleName  string
	Description string
}

var (
	ErrInvalidParams    = errors.New("invalid module params")
	ErrDBNotInitialized = errors.New("db pool is not initialized")
	ErrOwnerNotFound    = errors.New("owner user not found")
	ErrModuleNotFound   = errors.New("module not found")
)

type Service struct {
	db   *sql.DB
	repo *Repository
}

func NewService(database *sql.DB, repo *Repository) *Service {
	if repo == nil {
		repo = NewRepository(dao.NewACLDAO())
	}
	return &Service{db: database, repo: repo}
}

func (s *Service) CreateModule(params CreateModuleParams) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	moduleName := strings.TrimSpace(params.ModuleName)
	if moduleCode == "" || moduleName == "" {
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

	ownerUserID := params.UserID
	if ownerUserID > 0 {
		if err = s.repo.EnsureOwnerExists(tx, ownerUserID); err != nil {
			return 0, err
		}
	}

	moduleID, err := s.repo.InsertModule(tx, ownerUserID, moduleCode, moduleName, strings.TrimSpace(params.ModuleDescription))
	if err != nil {
		return 0, err
	}

	if ownerUserID > 0 {
		if err = s.repo.InsertModuleUser(tx, moduleID, ownerUserID); err != nil {
			return 0, err
		}
	}

	roleIDs, err := s.repo.InsertDefaultRoles(tx, moduleID, moduleCode)
	if err != nil {
		return 0, err
	}

	permIDs, err := s.repo.InsertDefaultPermissions(tx, moduleID)
	if err != nil {
		return 0, err
	}

	resIDs, err := s.repo.InsertDefaultResources(tx, moduleID, moduleCode)
	if err != nil {
		return 0, err
	}

	if ownerUserID > 0 {
		if err = s.repo.BindUserRole(tx, ownerUserID, moduleID, roleIDs[s.repo.ModuleRoleCode(moduleCode, "admin")]); err != nil {
			return 0, err
		}
	}

	if err = s.repo.GrantDefaultRolePermissions(tx, moduleID, roleIDs, permIDs, resIDs, moduleCode); err != nil {
		return 0, err
	}

	if ownerUserID > 0 {
		if err = s.repo.RebuildUserPermResByUser(tx, moduleID, ownerUserID); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return moduleID, nil
}

func (s *Service) SetModuleOwnerAndGrantAdmin(moduleCode string, ownerUserID int64) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || ownerUserID <= 0 {
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

	if err = s.repo.EnsureOwnerExists(tx, ownerUserID); err != nil {
		return err
	}

	moduleID, err := s.repo.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return err
	}

	if err = s.repo.UpdateModuleOwner(tx, moduleID, ownerUserID); err != nil {
		return err
	}

	if err = s.repo.InsertModuleUser(tx, moduleID, ownerUserID); err != nil {
		return err
	}

	superRoleID, err := s.repo.EnsureSuperRole(tx, moduleID, moduleCode)
	if err != nil {
		return err
	}

	adminRoleID, err := s.repo.GetRoleIDByCode(tx, moduleID, s.repo.ModuleRoleCode(moduleCode, "admin"))
	if err != nil {
		return err
	}

	if err = s.repo.BindUserRole(tx, ownerUserID, moduleID, adminRoleID); err != nil {
		return err
	}

	if err = s.repo.BindUserRole(tx, ownerUserID, moduleID, superRoleID); err != nil {
		return err
	}

	if err = s.repo.RebuildUserPermResByUser(tx, moduleID, ownerUserID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetByID(moduleID int64) (*Module, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}
	if moduleID <= 0 {
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

	m, err := s.repo.GetModuleByID(tx, moduleID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Service) GetList() ([]*Module, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
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

	list, err := s.repo.GetModuleList(tx)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Service) Update(params UpdateModuleParams) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}

	moduleName := strings.TrimSpace(params.ModuleName)
	if params.ModuleID <= 0 || moduleName == "" {
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

	if err = s.repo.UpdateModule(tx, params.ModuleID, moduleName, strings.TrimSpace(params.Description)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(moduleID int64) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if moduleID <= 0 {
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

	if err = s.repo.DeleteModule(tx, moduleID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
