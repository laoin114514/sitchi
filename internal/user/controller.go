package user

import (
	"net/http"
	"strconv"
	"strings"

	"sitchi/configs/db"
	"sitchi/internal/common/model"
	"sitchi/internal/dao"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController() *Controller {
	repo := NewRepository(dao.NewACLDAO())
	service := NewService(db.Pool, repo)
	return &Controller{service: service}
}

type LoginRequest struct {
	UserCode string `json:"user_code"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	ModuleCode  string `json:"module_code"`
	UserID      int64  `json:"user_id"`
	UserName    string `json:"user_name"`
	Description string `json:"description"`
	Email       string `json:"email"`
}

type DeleteUserRequest struct {
	ModuleCode string `json:"module_code"`
	UserID     int64  `json:"user_id"`
}

type CreateUserRequest struct {
	ModuleCode  string `json:"module_code"`
	UserCode    string `json:"user_code"`
	UserName    string `json:"user_name"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

type BindRoleRequest struct {
	ModuleCode string `json:"module_code"`
	UserID     int64  `json:"user_id"`
	RoleID     int64  `json:"role_id"`
}

// Login 登录接口：校验账号密码并签发 JWT（access/refresh）。
func (ctl *Controller) Login(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	result, err := ctl.service.Login(LoginParams{
		ModuleCode: "admin",
		UserCode:   strings.TrimSpace(req.UserCode),
		Password:   req.Password,
	})
	if err != nil {
		switch err {
		case ErrInvalidLoginParams:
			appErr := model.ErrInvalidParams.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		case ErrUserNotFound, ErrPasswordMismatch:
			appErr := model.ErrUnauthorized.WithDetail("用户名或密码错误")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		default:
			appErr := model.ErrInternal.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		}
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{
		"user_id":       result.UserID,
		"module_code":   result.ModuleCode,
		"roles":         result.Roles,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	}))
}

func (ctl *Controller) handleError(c *gin.Context, err error) {
	switch err {
	case ErrInvalidParams, ErrInvalidCreateUserParams:
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	case ErrModuleNotFound, ErrUserNotFound:
		appErr := model.ErrNotFound.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	default:
		appErr := model.ErrInternal.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	}
}

func (ctl *Controller) Create(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	userID, err := ctl.service.CreateUser(CreateUserParams{
		ModuleCode:  req.ModuleCode,
		UserCode:    req.UserCode,
		UserName:    req.UserName,
		Password:    req.Password,
		Email:       req.Email,
		Description: req.Description,
	})
	if err != nil {
		ctl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{"user_id": userID}))
}

func (ctl *Controller) GetByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if moduleCode == "" || err != nil || userID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("module_code and valid id are required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	user, svcErr := ctl.service.GetByID(moduleCode, userID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(user))
}

func (ctl *Controller) GetListByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	if moduleCode == "" {
		appErr := model.ErrInvalidParams.WithDetail("module_code is required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	list, err := ctl.service.GetList(moduleCode)
	if err != nil {
		ctl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(list))
}

func (ctl *Controller) Update(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Update(UpdateUserParams{
		ModuleCode:  req.ModuleCode,
		UserID:      req.UserID,
		UserName:    req.UserName,
		Description: req.Description,
		Email:       req.Email,
	})
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}

func (ctl *Controller) Delete(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Delete(req.ModuleCode, req.UserID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}

func (ctl *Controller) BindRole(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req BindRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.BindRole(req.ModuleCode, req.UserID, req.RoleID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}

func (ctl *Controller) UnbindRole(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("user controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req BindRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.UnbindRole(req.ModuleCode, req.UserID, req.RoleID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}
