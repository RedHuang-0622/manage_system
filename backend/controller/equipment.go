package controller

import (
	"strconv"

	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"manage_system/service"

	"github.com/gin-gonic/gin"
)

type EquipmentController struct {
	equipService *service.EquipmentService
}

func NewEquipmentController(equipService *service.EquipmentService) *EquipmentController {
	return &EquipmentController{equipService: equipService}
}

func (ctl *EquipmentController) List(c *gin.Context) {
	var req service.ListEquipReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	result, err := ctl.equipService.ListPage(c.Request.Context(), &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.PageResponse(c, result.Total, result.Page, result.PageSize, result.List)
}

func (ctl *EquipmentController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的设备ID")
		return
	}

	dto, err := ctl.equipService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, dto)
}

func (ctl *EquipmentController) Create(c *gin.Context) {
	var req service.CreateEquipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	equip, err := ctl.equipService.Create(c.Request.Context(), &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Created(c, equip)
}

func (ctl *EquipmentController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的设备ID")
		return
	}

	var req service.UpdateEquipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	if err := ctl.equipService.Update(c.Request.Context(), uint(id), &req); err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, gin.H{"msg": "更新成功"})
}

func (ctl *EquipmentController) Disable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的设备ID")
		return
	}

	if err := ctl.equipService.Disable(c.Request.Context(), uint(id)); err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	// 事务成功后失效缓存
	go ctl.equipService.InvalidateEquipmentCache(c.Request.Context(), uint(id))

	response.Success(c, gin.H{"msg": "设备已下架"})
}
