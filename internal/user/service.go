package user

import (
	"database/sql"
	"errors"
	"strings"

	"sitchi/internal/common"
	"sitchi/internal/dao"

	"golang.org/x/crypto/bcrypt"
)

type CreateUserParams struct {
	ModuleCode  string
	UserCode    string
	UserName    string
	Password    string
	Email       string
	Description string
}

type UpdateUserParams struct {
	ModuleCode  string
	UserID      int64
	UserName    string
	Description string
	Email       string
}

type LoginParams struct {
	ModuleCode string
	UserCode   string
	Password   string
}

type LoginResult struct {
	UserID       int64
	ModuleCode   string
	AccessToken  string
	RefreshToken string
	Roles        []string
}

var (
	ErrInvalidCreateUserParams = errors.New("invalid create user params")
	ErrInvalidLoginParams      = errors.New("invalid login params")
	ErrInvalidParams           = errors.New("invalid user params")
	ErrDBNotInitialized        = errors.New("db pool is not initialized")
	ErrModuleNotFound          = errors.New("module not found")
	ErrDefaultRoleNotFound     = errors.New("default role not found")
	ErrUserNotFound            = errors.New("user not found")
	ErrPasswordMismatch        = errors.New("password mismatch")
	ErrInvalidRefreshToken     = errors.New("invalid or expired refresh token")
)

type Service struct {
	db        *sql.DB
	repo      *Repository
	commonACL *common.ACLService
}

func NewService(database *sql.DB, repo *Repository) *Service {
	if repo == nil {
		repo = NewRepository(dao.NewACLDAO())
	}
	return &Service{db: database, repo: repo, commonACL: common.NewACLService(database, nil)}
}

func (s *Service) CreateUser(params CreateUserParams) (int64, error) {
	if s == nil || s.db == nil {
		return 0, ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	userCode := strings.TrimSpace(params.UserCode)
	userName := strings.TrimSpace(params.UserName)
	password := strings.TrimSpace(params.Password)

	if moduleCode == "" || userCode == "" || userName == "" || password == "" {
		return 0, ErrInvalidCreateUserParams
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	userID, err := s.repo.InsertUser(tx, moduleID, userCode, userName, strings.TrimSpace(params.Description), string(hashedPassword), strings.TrimSpace(params.Email))
	if err != nil {
		return 0, err
	}

	if err = s.repo.InsertModuleUser(tx, moduleID, userID, "模块用户"); err != nil {
		return 0, err
	}

	normalRoleCode := s.repo.ModuleRoleCode(moduleCode, "normal")
	roleID, err := s.repo.GetRoleIDByCode(tx, moduleID, normalRoleCode)
	if err != nil {
		return 0, err
	}

	if err = s.repo.BindUserRole(tx, userID, moduleID, roleID); err != nil {
		return 0, err
	}

	if err = s.repo.RebuildUserPermResByUser(tx, moduleID, userID); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *Service) Login(params LoginParams) (*LoginResult, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	userCode := strings.TrimSpace(params.UserCode)
	password := strings.TrimSpace(params.Password)
	if moduleCode == "" || userCode == "" || password == "" {
		return nil, ErrInvalidLoginParams
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

	rec, err := s.repo.GetUserAuthRecordByCode(tx, moduleCode, userCode)
	if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(rec.Password), []byte(password)) != nil {
		return nil, ErrPasswordMismatch
	}

	accessToken, err := common.JwtAuth.GenerateAccessToken(rec.UserID, rec.ModuleCode, rec.Roles)
	if err != nil {
		return nil, err
	}

	refreshToken, err := common.JwtAuth.GenerateRefreshToken(rec.UserID, rec.ModuleCode)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:       rec.UserID,
		ModuleCode:   rec.ModuleCode,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Roles:        rec.Roles,
	}, nil
}

func (s *Service) RefreshToken(refreshTokenStr string) (*LoginResult, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}

	claims, err := common.JwtAuth.ParseAndVerifyRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidRefreshToken
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

	moduleID, err := s.repo.GetModuleIDByCode(tx, claims.ModuleCode)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(tx, moduleID, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrUserNotFound
	}

	roles, err := s.repo.GetUserRolesByID(tx, moduleID, claims.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := common.JwtAuth.GenerateAccessToken(claims.UserID, claims.ModuleCode, roles)
	if err != nil {
		return nil, err
	}

	refreshToken, err := common.JwtAuth.GenerateRefreshToken(claims.UserID, claims.ModuleCode)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:       claims.UserID,
		ModuleCode:   claims.ModuleCode,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Roles:        roles,
	}, nil
}

func (s *Service) GetByID(moduleCode string, userID int64) (*User, error) {
	if s == nil || s.db == nil {
		return nil, ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || userID <= 0 {
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

	user, err := s.repo.GetUserByID(tx, moduleID, userID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetList(moduleCode string) ([]*User, error) {
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

	userList, err := s.repo.GetUserList(tx, moduleID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return userList, nil
}

func (s *Service) Update(params UpdateUserParams) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}

	moduleCode := strings.TrimSpace(params.ModuleCode)
	userName := strings.TrimSpace(params.UserName)
	if moduleCode == "" || userName == "" || params.UserID <= 0 {
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

	if err = s.repo.UpdateUser(tx, params.UserID, moduleID, userName, strings.TrimSpace(params.Description), strings.TrimSpace(params.Email)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(moduleCode string, userID int64) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || userID <= 0 {
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

	if err = s.repo.DeleteUser(tx, userID, moduleID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) BindRole(moduleCode string, userID, roleID int64) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || userID <= 0 || roleID <= 0 {
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

	if err = s.repo.BindUserRole(tx, userID, moduleID, roleID); err != nil {
		return err
	}

	if err = s.commonACL.RebuildUserPermResByUser(tx, moduleID, userID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) UnbindRole(moduleCode string, userID, roleID int64) error {
	if s == nil || s.db == nil {
		return ErrDBNotInitialized
	}
	if strings.TrimSpace(moduleCode) == "" || userID <= 0 || roleID <= 0 {
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

	if err = s.repo.UnbindUserRole(tx, userID, moduleID, roleID); err != nil {
		return err
	}

	if err = s.commonACL.RebuildUserPermResByUser(tx, moduleID, userID); err != nil {
		return err
	}

	return tx.Commit()
}
