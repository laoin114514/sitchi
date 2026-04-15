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
	ErrDBNotInitialized        = errors.New("db pool is not initialized")
	ErrModuleNotFound          = errors.New("module not found")
	ErrDefaultRoleNotFound     = errors.New("default role not found")
	ErrUserNotFound            = errors.New("user not found")
	ErrPasswordMismatch        = errors.New("password mismatch")
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
