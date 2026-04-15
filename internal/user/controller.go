package user

import (
	"net/http"
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
