package role

import (
	"net/http"
	"sitchi/configs/db"
	"sitchi/internal/common/model"
	"sitchi/internal/dao"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

type CreateRoleRequest struct {
	ModuleCode  string `json:"module_code"`
	RoleCode    string `json:"role_code"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	RoleCode    string `json:"role_code"`
	Description string `json:"description"`
}

func NewController() *Controller {
	repo := NewRepository(dao.NewACLDAO())
	service := NewService(db.Pool, repo)
	return &Controller{service: service}
}

// 统一错误处理
func (ctl *Controller) handleError(c *gin.Context, err error) {
	switch err {
	case ErrInvalidParams:
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	case ErrModuleNotFound, ErrRoleNotFound:
		appErr := model.ErrNotFound.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	default:
		appErr := model.ErrInternal.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	}
}

func (ctl *Controller) Create(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("role controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	roleID, err := ctl.service.CreateRole(CreateRoleParams{
		ModuleCode:  req.ModuleCode,
		RoleCode:    req.RoleCode,
		Description: req.Description,
	})
	if err != nil {
		ctl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(roleID))
}

func (ctl *Controller) GetByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("role controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if moduleCode == "" || err != nil || roleID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("module_code and valid id are required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	role, svcErr := ctl.service.GetByID(moduleCode, roleID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(role))
}

func (ctl *Controller) GetListByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("role controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	if moduleCode == "" {
		appErr := model.ErrInvalidParams.WithDetail("module_code is required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	list, err := ctl.service.GetListByID(moduleCode)
	if err != nil {
		ctl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(list))
}

func (ctl *Controller) Update(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("role controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if moduleCode == "" || err != nil || roleID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("module_code and valid id are required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail("invalid request body")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Update(UpdateRoleParams{
		ModuleCode:  moduleCode,
		RoleID:      roleID,
		RoleCode:    req.RoleCode,
		Description: req.Description,
	})
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}

func (ctl *Controller) Delete(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("role controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if moduleCode == "" || err != nil || roleID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("module_code and valid id are required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Delete(moduleCode, roleID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}
