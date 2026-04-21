package module

import (
	"database/sql"
	"errors"
	"strings"

	"sitchi/internal/dao"
)

// GetModulesParams 查询模块列表参数
type GetModulesParams struct {
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Filters  map[string]string `json:"filters"`
}

// UpdateModuleParams 更新模块参数
type UpdateModuleParams struct {
	ModuleID    int64  `json:"module_id"`
	ModuleName  string `json:"module_name"`
	Description string `json:"description"`
}

type CreateModuleParams struct {
	UserID            int64
	ModuleCode        string
	ModuleName        string
	ModuleDescription string
}

var (
	ErrInvalidParams    = errors.New("invalid create module params")
	ErrDBNotInitialized = errors.New("db pool is not initialized")
	ErrOwnerNotFound    = errors.New("owner user not found")
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

func (s *Service) GetModules(params GetModulesParams) ([]Module, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, ErrDBNotInitialized
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 10
	}

	// 转换过滤条件
	filters := make(map[string]interface{})
	if params.Filters != nil {
		for k, v := range params.Filters {
			filters[k] = v
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	modules, total, err := s.repo.GetModules(tx, params.Page, params.PageSize, filters)
	if err != nil {
		return nil, 0, err
	}

	return modules, total, nil
}

func (s *Service) GetModuleByID(moduleID int64) (Module, error) {
	if s == nil || s.db == nil {
		return Module{}, ErrDBNotInitialized
	}
	if moduleID <= 0 {
		return Module{}, ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Module{}, err
	}
	defer tx.Rollback()

	module, err := s.repo.GetModuleByID(tx, moduleID)
	if err != nil {
		return Module{}, err
	}

	return module, nil
}

func (s *Service) GetModuleByCode(moduleCode string) (Module, error) {
	if s == nil || s.db == nil {
		return Module{}, ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" {
		return Module{}, ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Module{}, err
	}
	defer tx.Rollback()

	module, err := s.repo.GetModuleByCode(tx, moduleCode)
	if err != nil {
		return Module{}, err
	}

	return module, nil
}

func (s *Service) UpdateModule(params UpdateModuleParams) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if params.ModuleID <= 0 || strings.TrimSpace(params.ModuleName) == "" {
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

	// 验证模块是否存在
	_, err = s.repo.GetModuleByID(tx, params.ModuleID)
	if err != nil {
		return err
	}

	// 更新模块信息
	err = s.repo.UpdateModule(tx, params.ModuleID, strings.TrimSpace(params.ModuleName), strings.TrimSpace(params.Description))
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteModule(moduleID int64) error {
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

	// 验证模块是否存在
	_, err = s.repo.GetModuleByID(tx, moduleID)
	if err != nil {
		return err
	}

	// 软删除模块
	err = s.repo.DeleteModule(tx, moduleID)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
