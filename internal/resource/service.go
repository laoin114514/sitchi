package resource

import (
	"database/sql"
	"errors"
	"strings"

	"sitchi/configs/db"
	"sitchi/internal/dao"
)

// 错误定义
var (
	ErrDBNotInitialized     = errors.New("db pool is not initialized")
	ErrInvalidParams        = errors.New("invalid parameters")
	ErrResourceNotFound     = errors.New("resource not found")
	ErrResourceCodeExists   = errors.New("resource code already exists")
	ErrResourceCodeRequired = errors.New("resource code is required")
	ErrResourceNameRequired = errors.New("resource name is required")
	ErrResourceTypeRequired = errors.New("resource type is required")
)

// CreateResourceParams 创建资源参数
type CreateResourceParams struct {
	ModuleCode  string
	ResCode     string
	ResName     string
	ResType     string
	ParentID    *int64
	Path        string
	Description string
}

// UpdateResourceParams 更新资源参数
type UpdateResourceParams struct {
	ResourceID  int64
	ModuleCode  string
	ResName     string
	ResType     string
	ParentID    *int64
	Path        string
	Description string
}

// Service 资源服务层
type Service struct {
	db   *sql.DB
	repo *Repository
}

// NewService 创建 Service 实例
func NewService(database *sql.DB, repo *Repository) *Service {
	if repo == nil {
		repo = NewRepository(dao.NewACLDAO())
	}
	return &Service{db: database, repo: repo}
}

// Create 创建资源
func (s *Service) Create(params CreateResourceParams) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrDBNotInitialized
	}

	// 参数校验
	resCode := strings.TrimSpace(params.ResCode)
	resName := strings.TrimSpace(params.ResName)
	resType := strings.TrimSpace(params.ResType)

	if resCode == "" {
		return 0, ErrResourceCodeRequired
	}
	if resName == "" {
		return 0, ErrResourceNameRequired
	}
	if resType == "" {
		return 0, ErrResourceTypeRequired
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

	// 获取模块ID
	moduleID, err := s.repo.aclDAO.GetModuleIDByCode(tx, params.ModuleCode)
	if err != nil {
		return 0, err
	}

	// 检查资源编码是否已存在
	exists, err := s.repo.CheckCodeExists(tx, moduleID, resCode)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrResourceCodeExists
	}

	// 创建资源
	resourceID, err := s.repo.Insert(tx, moduleID, resCode, resName, resType, params.ParentID, params.Path, params.Description)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return resourceID, nil
}

// Update 更新资源
func (s *Service) Update(params UpdateResourceParams) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}

	// 参数校验
	resName := strings.TrimSpace(params.ResName)
	resType := strings.TrimSpace(params.ResType)

	if params.ResourceID <= 0 {
		return ErrInvalidParams
	}
	if resName == "" {
		return ErrResourceNameRequired
	}
	if resType == "" {
		return ErrResourceTypeRequired
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

	// 获取模块ID
	moduleID, err := s.repo.aclDAO.GetModuleIDByCode(tx, params.ModuleCode)
	if err != nil {
		return err
	}

	// 检查资源是否存在
	resource, err := s.repo.GetByID(tx, params.ResourceID, moduleID)
	if err != nil {
		return err
	}
	if resource == nil {
		return ErrResourceNotFound
	}

	// 更新资源
	err = s.repo.Update(tx, params.ResourceID, moduleID, resName, resType, params.ParentID, params.Path, params.Description)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete 删除资源
func (s *Service) Delete(resourceID int64, moduleCode string) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}

	if resourceID <= 0 {
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

	// 获取模块ID
	moduleID, err := s.repo.aclDAO.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return err
	}

	// 检查资源是否存在
	resource, err := s.repo.GetByID(tx, resourceID, moduleID)
	if err != nil {
		return err
	}
	if resource == nil {
		return ErrResourceNotFound
	}

	// 删除资源
	err = s.repo.Delete(tx, resourceID, moduleID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetByID 根据ID获取资源
func (s *Service) GetByID(resourceID int64, moduleCode string) (*Resource, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}

	if resourceID <= 0 {
		return nil, ErrInvalidParams
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 获取模块ID
	moduleID, err := s.repo.aclDAO.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return nil, err
	}

	resource, err := s.repo.GetByID(tx, resourceID, moduleID)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, ErrResourceNotFound
	}

	return resource, nil
}

// List 获取资源列表
func (s *Service) List(moduleCode string, resType string, parentID *int64) ([]*Resource, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 获取模块ID
	moduleID, err := s.repo.aclDAO.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return nil, err
	}

	resources, err := s.repo.List(tx, moduleID, resType, parentID)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

// GetDB 获取数据库连接（供 Controller 使用）
func (s *Service) GetDB() *sql.DB {
	if s == nil {
		return db.Pool
	}
	return s.db
}
