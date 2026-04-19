package perm

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

type CreatePermRequest struct {
	ModuleCode  string `json:"module_code"`
	PermCode    string `json:"perm_code"`
	PermName    string `json:"perm_name"`
	Description string `json:"description"`
}

type UpdatePermRequest struct {
	PermCode    string `json:"perm_code"`
	PermName    string `json:"perm_name"`
	Description string `json:"description"`
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
	case ErrModuleNotFound, ErrPermNotFound:
		appErr := model.ErrNotFound.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	default:
		appErr := model.ErrInternal.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
	}
}

func (ctl *Controller) Create(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("perm controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req CreatePermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	permID, err := ctl.service.CreatePerm(CreatePermParams{
		ModuleCode:  req.ModuleCode,
		PermCode:    req.PermCode,
		PermName:    req.PermName,
		Description: req.Description,
	})
	if err != nil {
		ctl.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{"perm_id": permID}))
}

func (ctl *Controller) GetByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("perm controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	permID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if moduleCode == "" || err != nil || permID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("module_code and valid id are required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	perm, svcErr := ctl.service.GetByID(moduleCode, permID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(perm))
}

func (ctl *Controller) GetListByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("perm controller not initialized")
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
		appErr := model.ErrInternal.WithDetail("perm controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	permID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if moduleCode == "" || err != nil || permID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("module_code and valid id are required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req UpdatePermRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		appErr := model.ErrInvalidParams.WithDetail(bindErr.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Update(UpdatePermParams{
		ModuleCode:  moduleCode,
		PermID:      permID,
		PermCode:    req.PermCode,
		PermName:    req.PermName,
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
		appErr := model.ErrInternal.WithDetail("perm controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	moduleCode := strings.TrimSpace(c.Query("module_code"))
	permID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if moduleCode == "" || err != nil || permID <= 0 {
		appErr := model.ErrInvalidParams.WithDetail("module_code and valid id are required")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	svcErr := ctl.service.Delete(moduleCode, permID)
	if svcErr != nil {
		ctl.handleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}
