package common

import (
	"database/sql"
	"errors"

	"sitchi/internal/dao"
)

// ACLService 提供跨模块复用的 ACL 通用操作。
type ACLService struct {
	db     *sql.DB
	aclDAO *dao.ACLDAO
}

func NewACLService(db *sql.DB, aclDAO *dao.ACLDAO) *ACLService {
	if aclDAO == nil {
		aclDAO = dao.NewACLDAO()
	}
	return &ACLService{db: db, aclDAO: aclDAO}
}

// RebuildUserPermResByUser 重建指定用户在指定模块下的权限资源缓存。
// 接受 *sql.Tx 以便调用方组合进自己的事务中。
func (s *ACLService) RebuildUserPermResByUser(tx *sql.Tx, moduleID, userID int64) error {
	return s.aclDAO.RebuildUserPermResByUser(tx, moduleID, userID)
}

// BindRolePermRes 为角色授予(权限,资源)对，事务内完成写入+缓存刷新。
func (s *ACLService) BindRolePermRes(moduleCode string, roleID, permID, resID int64) error {
	if s == nil || s.db == nil {
		return ErrACLDBNotInitialized
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

	moduleID, err := s.aclDAO.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return err
	}

	if err = s.aclDAO.BindRolePermRes(tx, moduleID, roleID, permID, resID); err != nil {
		return err
	}

	if err = s.aclDAO.RebuildUserPermResByRole(tx, moduleID, roleID); err != nil {
		return err
	}

	return tx.Commit()
}

// UnbindRolePermRes 解除角色对(权限,资源)的绑定，事务内完成删除+缓存刷新。
func (s *ACLService) UnbindRolePermRes(moduleCode string, roleID, permID, resID int64) error {
	if s == nil || s.db == nil {
		return ErrACLDBNotInitialized
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

	moduleID, err := s.aclDAO.GetModuleIDByCode(tx, moduleCode)
	if err != nil {
		return err
	}

	if err = s.aclDAO.UnbindRolePermRes(tx, moduleID, roleID, permID, resID); err != nil {
		return err
	}

	if err = s.aclDAO.RebuildUserPermResByRole(tx, moduleID, roleID); err != nil {
		return err
	}

	return tx.Commit()
}

var (
	ErrACLDBNotInitialized = errors.New("db pool is not initialized")
)
