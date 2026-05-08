package common

import (
	"database/sql"

	"sitchi/internal/dao"
)

// ACLService 提供跨模块复用的 ACL 通用操作。
type ACLService struct {
	aclDAO *dao.ACLDAO
}

func NewACLService(aclDAO *dao.ACLDAO) *ACLService {
	if aclDAO == nil {
		aclDAO = dao.NewACLDAO()
	}
	return &ACLService{aclDAO: aclDAO}
}

// RebuildUserPermResByUser 重建指定用户在指定模块下的权限资源缓存。
// 接受 *sql.Tx 以便调用方组合进自己的事务中。
func (s *ACLService) RebuildUserPermResByUser(tx *sql.Tx, moduleID, userID int64) error {
	return s.aclDAO.RebuildUserPermResByUser(tx, moduleID, userID)
}
