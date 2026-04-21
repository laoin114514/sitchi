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

func NewController() *Controller {
	repo := NewRepository(dao.NewACLDAO())
	service := NewService(db.Pool, repo)
	return &Controller{service: service}
}

// GetModules 获取模块列表
// @Summary 获取模块列表
// @Description 分页获取模块列表，支持按模块名称和代码筛选
// @Tags 模块管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页条数，默认10"
// @Param module_name query string false "模块名称"
// @Param module_code query string false "模块代码"
// @Success 200 {object} map[string]interface{}
// @Router /module/list [get]
func (c *Controller) GetModules(ctx *gin.Context) {
	// 解析参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	moduleName := ctx.Query("module_name")
	moduleCode := ctx.Query("module_code")

	// 构建查询参数
	params := GetModulesParams{
		Page:     page,
		PageSize: pageSize,
		Filters: map[string]string{
			"module_name": moduleName,
			"module_code": moduleCode,
		},
	}

	// 调用服务层方法
	modules, total, err := c.service.GetModules(params)
	if err != nil {
		appErr := model.ErrInternal.WithDetail("获取模块列表失败: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	data := gin.H{
		"list":  modules,
		"total": total,
		"page":  page,
		"size":  pageSize,
	}
	ctx.JSON(http.StatusOK, model.ApiSuccessResponse(data))
}

// GetModuleDetail 获取模块详情
// @Summary 获取模块详情
// @Description 根据模块ID或模块代码获取模块详情
// @Tags 模块管理
// @Accept json
// @Produce json
// @Param id query int64 false "模块ID"
// @Param code query string false "模块代码"
// @Success 200 {object} map[string]interface{}
// @Router /module/detail [get]
func (c *Controller) GetModuleDetail(ctx *gin.Context) {
	// 解析参数
	idStr := ctx.Query("id")
	code := ctx.Query("code")

	var module Module
	var err error

	// 根据ID或代码查询
	if idStr != "" {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		module, err = c.service.GetModuleByID(id)
	} else if code != "" {
		module, err = c.service.GetModuleByCode(code)
	} else {
		appErr := model.ErrInvalidParams.WithDetail("必须提供模块ID或模块代码")
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	if err != nil {
		appErr := model.ErrInternal.WithDetail("获取模块详情失败: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	ctx.JSON(http.StatusOK, model.ApiSuccessResponse(module))
}

// CreateModule 创建模块
// @Summary 创建模块
// @Description 创建新的模块
// @Tags 模块管理
// @Accept json
// @Produce json
// @Param data body CreateModuleParams true "模块信息"
// @Success 200 {object} map[string]interface{}
// @Router /module/create [post]
func (c *Controller) CreateModule(ctx *gin.Context) {
	// 解析参数
	var params CreateModuleParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		appErr := model.ErrInvalidParams.WithDetail("参数错误: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	// 调用服务层方法
	moduleID, err := c.service.CreateModule(params)
	if err != nil {
		appErr := model.ErrInternal.WithDetail("创建模块失败: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	data := gin.H{
		"module_id": moduleID,
	}
	ctx.JSON(http.StatusOK, model.ApiSuccessResponse(data))
}

// UpdateModule 更新模块
// @Summary 更新模块
// @Description 更新模块信息
// @Tags 模块管理
// @Accept json
// @Produce json
// @Param data body UpdateModuleParams true "模块信息"
// @Success 200 {object} map[string]interface{}
// @Router /module/update [post]
func (c *Controller) UpdateModule(ctx *gin.Context) {
	// 解析参数
	var params UpdateModuleParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		appErr := model.ErrInvalidParams.WithDetail("参数错误: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	// 调用服务层方法
	err := c.service.UpdateModule(params)
	if err != nil {
		appErr := model.ErrInternal.WithDetail("更新模块失败: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	ctx.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}

// DeleteModule 删除模块
// @Summary 删除模块
// @Description 软删除模块
// @Tags 模块管理
// @Accept json
// @Produce json
// @Param data body map[string]int64 true "模块ID"
// @Success 200 {object} map[string]interface{}
// @Router /module/delete [post]
func (c *Controller) DeleteModule(ctx *gin.Context) {
	// 解析参数
	var params struct {
		ModuleID int64 `json:"module_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&params); err != nil {
		appErr := model.ErrInvalidParams.WithDetail("参数错误: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	// 调用服务层方法
	err := c.service.DeleteModule(params.ModuleID)
	if err != nil {
		appErr := model.ErrInternal.WithDetail("删除模块失败: " + err.Error())
		ctx.JSON(appErr.HTTPStatus(), model.ApiErrorResponse(appErr.Code, appErr.Message, appErr))
		return
	}

	ctx.JSON(http.StatusOK, model.ApiSuccessResponse(nil))
}
