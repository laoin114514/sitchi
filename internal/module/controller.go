package module

import (
	"net/http"
	"strconv"

	"sitchi/configs/db"
	"sitchi/internal/common/model"
	"sitchi/internal/dao"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

type CreateModuleRequest struct {
	ModuleCode        string `json:"module_code"`
	ModuleName        string `json:"module_name"`
	Description       string `json:"description"`
	OwnerUserID       int64  `json:"owner_user_id"`
}

type UpdateModuleRequest struct {
	ModuleID    int64  `json:"module_id"`
	ModuleName  string `json:"module_name"`
	Description string `json:"description"`
}

type DeleteModuleRequest struct {
	ModuleID int64 `json:"module_id"`
}

func NewController() *Controller {
	repo := NewRepository(dao.NewACLDAO())
	service := NewService(db.Pool, repo)
	return &Controller{service: service}
}

func (ctl *Controller) handleError(c *gin.Context, err error) {
	switch err {
	case ErrInvalidParams:
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	case ErrModuleNotFound, ErrOwnerNotFound:
		appErr := model.ErrNotFound.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	default:
		appErr := model.ErrInternal.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	}
}

// Create 创建模块并初始化默认角色/权限/资源。
func (ctl *Controller) Create(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("module controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleID, err := ctl.service.CreateModule(CreateModuleParams{
		UserID:            req.OwnerUserID,
		ModuleCode:        req.ModuleCode,
		ModuleName:        req.ModuleName,
		ModuleDescription: req.Description,
	})
	if err != nil {
		ctl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{"module_id": moduleID}))
}

func (ctl *Controller) GetByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("module controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || moduleID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("valid id is required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	m, svcErr := ctl.service.GetByID(moduleID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(m))
}

func (ctl *Controller) GetList(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("module controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	list, err := ctl.service.GetList()
	if err != nil {
		ctl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(list))
}

func (ctl *Controller) Update(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("module controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Update(UpdateModuleParams{
		ModuleID:    req.ModuleID,
		ModuleName:  req.ModuleName,
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
		appErr := model.ErrInternal.WithDetail("module controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req DeleteModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Delete(req.ModuleID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}
