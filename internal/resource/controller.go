package resource

import (
	"net/http"
	"strconv"

	"sitchi/configs/db"
	"sitchi/internal/common/model"

	"github.com/gin-gonic/gin"
)

// Controller 资源控制器
type Controller struct {
	service *Service
}

// NewController 创建 Controller 实例
func NewController() *Controller {
	repo := NewRepository(nil)
	service := NewService(db.Pool, repo)
	return &Controller{service: service}
}

// CreateRequest 创建资源请求
type CreateRequest struct {
	ModuleCode  string `json:"module_code" binding:"required"`
	ResCode     string `json:"res_code" binding:"required"`
	ResName     string `json:"res_name" binding:"required"`
	ResType     string `json:"res_type" binding:"required"`
	ParentID    *int64 `json:"parent_id"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

// UpdateRequest 更新资源请求
type UpdateRequest struct {
	ModuleCode  string `json:"module_code" binding:"required"`
	ResourceID  int64  `json:"resource_id" binding:"required"`
	ResName     string `json:"res_name" binding:"required"`
	ResType     string `json:"res_type" binding:"required"`
	ParentID    *int64 `json:"parent_id"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

// DeleteRequest 删除资源请求
type DeleteRequest struct {
	ModuleCode string `json:"module_code" binding:"required"`
	ResourceID int64  `json:"resource_id" binding:"required"`
}

// GetByIDRequest 获取资源详情请求
type GetByIDRequest struct {
	ModuleCode string `json:"module_code" binding:"required"`
	ResourceID int64  `json:"resource_id" binding:"required"`
}

// Create 创建资源接口
// POST /api/webui/resources/create
func (ctl *Controller) Create(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("resource controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	resourceID, err := ctl.service.Create(CreateResourceParams{
		ModuleCode:  req.ModuleCode,
		ResCode:     req.ResCode,
		ResName:     req.ResName,
		ResType:     req.ResType,
		ParentID:    req.ParentID,
		Path:        req.Path,
		Description: req.Description,
	})
	if err != nil {
		switch err {
		case ErrResourceCodeRequired, ErrResourceNameRequired, ErrResourceTypeRequired, ErrInvalidParams:
			appErr := model.ErrInvalidParams.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		case ErrResourceCodeExists:
			appErr := model.ErrConflict.WithDetail("资源编码已存在")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		default:
			appErr := model.ErrInternal.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		}
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{
		"resource_id": resourceID,
		"message":     "资源创建成功",
	}))
}

// Update 更新资源接口
// POST /api/webui/resources/update
func (ctl *Controller) Update(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("resource controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	err := ctl.service.Update(UpdateResourceParams{
		ResourceID:  req.ResourceID,
		ModuleCode:  req.ModuleCode,
		ResName:     req.ResName,
		ResType:     req.ResType,
		ParentID:    req.ParentID,
		Path:        req.Path,
		Description: req.Description,
	})
	if err != nil {
		switch err {
		case ErrInvalidParams, ErrResourceNameRequired, ErrResourceTypeRequired:
			appErr := model.ErrInvalidParams.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		case ErrResourceNotFound:
			appErr := model.ErrNotFound.WithDetail("资源不存在")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		default:
			appErr := model.ErrInternal.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		}
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{
		"message": "资源更新成功",
	}))
}

// Delete 删除资源接口
// POST /api/webui/resources/delete
func (ctl *Controller) Delete(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("resource controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	err := ctl.service.Delete(req.ResourceID, req.ModuleCode)
	if err != nil {
		switch err {
		case ErrInvalidParams:
			appErr := model.ErrInvalidParams.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		case ErrResourceNotFound:
			appErr := model.ErrNotFound.WithDetail("资源不存在")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		default:
			appErr := model.ErrInternal.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		}
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{
		"message": "资源删除成功",
	}))
}

// GetByID 获取资源详情接口
// POST /api/webui/resources/detail
func (ctl *Controller) GetByID(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("resource controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	var req GetByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := model.ErrInvalidParams.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	resource, err := ctl.service.GetByID(req.ResourceID, req.ModuleCode)
	if err != nil {
		switch err {
		case ErrInvalidParams:
			appErr := model.ErrInvalidParams.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		case ErrResourceNotFound:
			appErr := model.ErrNotFound.WithDetail("资源不存在")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		default:
			appErr := model.ErrInternal.WithDetail(err.Error())
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		}
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(resource))
}

// List 获取资源列表接口
// GET /api/webui/resources/list
func (ctl *Controller) List(c *gin.Context) {
	if ctl == nil || ctl.service == nil {
		appErr := model.ErrInternal.WithDetail("resource controller not initialized")
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	// 获取查询参数
	resType := c.Query("res_type")
	parentIDStr := c.Query("parent_id")

	var parentID *int64
	if parentIDStr != "" {
		pid, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			appErr := model.ErrInvalidParams.WithDetail("无效的父资源ID")
			c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
			return
		}
		parentID = &pid
	}

	moduleCode := c.Query("module_code")
	if moduleCode == "" {
		moduleCode = "admin"
	}

	resources, err := ctl.service.List(moduleCode, resType, parentID)
	if err != nil {
		appErr := model.ErrInternal.WithDetail(err.Error())
		c.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	c.JSON(http.StatusOK, model.ApiSuccessResponse(gin.H{
		"list":  resources,
		"total": len(resources),
	}))
}
