package controller

import (
	"strconv"

	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"manage_system/service"

	"github.com/gin-gonic/gin"
)

type BorrowController struct {
	borrowService *service.BorrowService
}

func NewBorrowController(borrowService *service.BorrowService) *BorrowController {
	return &BorrowController{borrowService: borrowService}
}

func (ctl *BorrowController) Apply(c *gin.Context) {
	var req service.ApplyBorrowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint("user_id")
	record, err := ctl.borrowService.Apply(c.Request.Context(), userID, &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Created(c, record)
}

func (ctl *BorrowController) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的工单ID")
		return
	}

	var req service.ApproveBorrowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	approverID := c.GetUint("user_id")
	record, err := ctl.borrowService.Approve(c.Request.Context(), uint(id), req.Approve, approverID, req.ApproveNote)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, record)
}

func (ctl *BorrowController) Return(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的工单ID")
		return
	}

	operatorID := c.GetUint("user_id")
	isAdmin := isAdminRole(c.GetString("role_name"))

	record, err := ctl.borrowService.Return(c.Request.Context(), uint(id), operatorID, isAdmin)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, record)
}

func (ctl *BorrowController) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的工单ID")
		return
	}

	operatorID := c.GetUint("user_id")
	if err := ctl.borrowService.Cancel(c.Request.Context(), uint(id), operatorID); err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, gin.H{"msg": "已取消申请"})
}

func (ctl *BorrowController) ListMyRecords(c *gin.Context) {
	var req service.ListBorrowReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	userID := c.GetUint("user_id")
	result, err := ctl.borrowService.ListMyRecords(c.Request.Context(), userID, &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.PageResponse(c, result.Total, result.Page, result.PageSize, result.List)
}

func (ctl *BorrowController) ListPending(c *gin.Context) {
	var req service.ListBorrowReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	result, err := ctl.borrowService.ListPending(c.Request.Context(), &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.PageResponse(c, result.Total, result.Page, result.PageSize, result.List)
}

func (ctl *BorrowController) ListAll(c *gin.Context) {
	var req service.ListBorrowReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	result, err := ctl.borrowService.ListAll(c.Request.Context(), &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.PageResponse(c, result.Total, result.Page, result.PageSize, result.List)
}

// ──────────────────────────── 工具函数 ────────────────────────────

func isAdminRole(roleName string) bool {
	return roleName == "super_admin" || roleName == "lab_admin"
}
