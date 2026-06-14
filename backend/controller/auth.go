package controller

import (
	"strconv"

	"manage_system/pkg/errcode"
	"manage_system/pkg/response"
	"manage_system/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	iamService *service.IAMService
}

func NewAuthController(iamService *service.IAMService) *AuthController {
	return &AuthController{iamService: iamService}
}

// Login 登录
func (ctl *AuthController) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	resp, err := ctl.iamService.Login(c.Request.Context(), &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, resp)
}

// Logout 登出
func (ctl *AuthController) Logout(c *gin.Context) {
	tokenStr := extractToken(c)
	if tokenStr == "" {
		response.Error(c, 401, errcode.ErrTokenMissing, "未提供认证Token")
		return
	}

	_ = ctl.iamService.Logout(c.Request.Context(), tokenStr)
	response.Success(c, gin.H{"msg": "登出成功"})
}

// RefreshToken 刷新Token
func (ctl *AuthController) RefreshToken(c *gin.Context) {
	tokenStr := extractToken(c)
	if tokenStr == "" {
		response.Error(c, 401, errcode.ErrTokenMissing, "未提供认证Token")
		return
	}

	resp, err := ctl.iamService.RefreshToken(c.Request.Context(), tokenStr)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, resp)
}

// ──────────────────────────── 角色 ────────────────────────────

func (ctl *AuthController) ListRoles(c *gin.Context) {
	roles, err := ctl.iamService.ListRoles(c.Request.Context())
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}
	response.Success(c, roles)
}

// ──────────────────────────── 用户管理 ────────────────────────────

func (ctl *AuthController) CreateUser(c *gin.Context) {
	var req service.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	user, err := ctl.iamService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Created(c, user)
}

func (ctl *AuthController) ListUsers(c *gin.Context) {
	var req service.ListUserReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	result, err := ctl.iamService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.PageResponse(c, result.Total, result.Page, result.PageSize, result.List)
}

func (ctl *AuthController) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的用户ID")
		return
	}

	operatorID := c.GetUint("user_id")
	operatorRole := c.GetString("role_name")

	dto, err := ctl.iamService.GetUserByID(c.Request.Context(), uint(id), operatorID, operatorRole)
	if err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, dto)
}

func (ctl *AuthController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的用户ID")
		return
	}

	var req service.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	if err := ctl.iamService.UpdateUser(c.Request.Context(), uint(id), &req); err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, gin.H{"msg": "更新成功"})
}

func (ctl *AuthController) DisableUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的用户ID")
		return
	}

	operatorID := c.GetUint("user_id")
	if err := ctl.iamService.DisableUser(c.Request.Context(), uint(id), operatorID); err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, gin.H{"msg": "用户已禁用"})
}

func (ctl *AuthController) ChangePassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "无效的用户ID")
		return
	}

	var req service.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, errcode.ErrInvalidParam, "请求参数错误: "+err.Error())
		return
	}

	operatorID := c.GetUint("user_id")
	isAdmin := c.GetString("role_name") == "super_admin"

	if err := ctl.iamService.ChangePassword(c.Request.Context(), operatorID, uint(id), &req, isAdmin); err != nil {
		code, msg := parseServiceError(err)
		response.Error(c, codeToHTTP(code), code, msg)
		return
	}

	response.Success(c, gin.H{"msg": "密码修改成功"})
}
