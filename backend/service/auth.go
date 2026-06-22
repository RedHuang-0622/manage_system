package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"manage_system/dao"
	"manage_system/models"

	"manage_system/pkg/errcode"
	jwtpkg "manage_system/pkg/jwt"
	"manage_system/pkg/response"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IAMService struct {
	db          *gorm.DB
	userDAO     dao.UserDAO
	roleDAO     dao.RoleDAO
	jwtService  *jwtpkg.Service
	redisClient *redis.Client
}

func NewIAMService(db *gorm.DB, userDAO dao.UserDAO, roleDAO dao.RoleDAO, jwtService *jwtpkg.Service, rdb *redis.Client) *IAMService {
	return &IAMService{
		db:          db,
		userDAO:     userDAO,
		roleDAO:     roleDAO,
		jwtService:  jwtService,
		redisClient: rdb,
	}
}

// ──────────────────────────── 认证 ────────────────────────────

type LoginReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

type LoginResp struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

type RoleDTO struct {
	ID          uint   `json:"id"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
}

func (s *IAMService) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	user, err := s.userDAO.FindByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if user == nil {
		return nil, fmt.Errorf("[%d] 用户名或密码错误", errcode.ErrAuthFailed)
	}
	if user.Status == 0 {
		return nil, fmt.Errorf("[%d] 账号已被禁用", errcode.ErrAccountDisabled)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("[%d] 用户名或密码错误", errcode.ErrAuthFailed)
	}

	role, err := s.roleDAO.FindByID(user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	roleName := ""
	if role != nil {
		roleName = role.RoleName
	}

	token, expiresIn, err := s.jwtService.GenerateToken(user.ID, user.Username, user.RoleID, roleName)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	return &LoginResp{Token: token, ExpiresIn: expiresIn}, nil
}

func (s *IAMService) Logout(ctx context.Context, tokenStr string) error {
	claims, err := s.jwtService.ParseToken(tokenStr)
	if err != nil {
		return nil // 即使解析失败也视为登出成功（Token可能已过期）
	}
	if claims.ExpiresAt != nil {
		return s.jwtService.AddToBlacklist(tokenStr, claims.ExpiresAt.Time)
	}
	return nil
}

func (s *IAMService) RefreshToken(ctx context.Context, oldToken string) (*LoginResp, error) {
	// 验证旧Token
	claims, err := s.jwtService.ParseToken(oldToken)
	if err != nil {
		return nil, fmt.Errorf("[%d] Token无效或已过期", errcode.ErrTokenInvalid)
	}
	if s.jwtService.IsInBlacklist(oldToken) {
		return nil, fmt.Errorf("[%d] Token已失效", errcode.ErrTokenInvalid)
	}

	// 旧Token加入黑名单
	if claims.ExpiresAt != nil {
		_ = s.jwtService.AddToBlacklist(oldToken, claims.ExpiresAt.Time)
	}

	// 签发新Token
	token, expiresIn, err := s.jwtService.GenerateToken(claims.UserID, claims.Username, claims.RoleID, claims.RoleName)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	return &LoginResp{Token: token, ExpiresIn: expiresIn}, nil
}

// ──────────────────────────── 角色 ────────────────────────────

func (s *IAMService) ListRoles(ctx context.Context) ([]RoleDTO, error) {
	roles, err := s.roleDAO.FindAll()
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	result := make([]RoleDTO, len(roles))
	for i, r := range roles {
		result[i] = RoleDTO{ID: r.ID, RoleName: r.RoleName, Description: r.Description}
	}
	return result, nil
}

// ──────────────────────────── 用户管理 ────────────────────────────

type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	RealName string `json:"real_name" binding:"required,min=2,max=32"`
	Email    string `json:"email" binding:"omitempty,email,max=64"`
	Phone    string `json:"phone" binding:"omitempty,numeric,max=16"`
	RoleID   uint   `json:"role_id" binding:"required,gt=0"`
}

type UpdateUserReq struct {
	RealName string `json:"real_name" binding:"omitempty,min=2,max=32"`
	Email    string `json:"email" binding:"omitempty,email,max=64"`
	Phone    string `json:"phone" binding:"omitempty,numeric,max=16"`
	RoleID   uint   `json:"role_id" binding:"omitempty"`
	Status   *int8  `json:"status"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

type ListUserReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	Status   *int   `form:"status"` // 指针类型：nil 表示不过滤，避免 int 零值误过滤 status=0
	RoleID   uint   `form:"role_id"`
}

type UserDTO struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	RealName  string    `json:"real_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	RoleID    uint      `json:"role_id"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Role      *RoleDTO  `json:"role,omitempty"`
}

type PageResult struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     interface{} `json:"list"`
}

func (s *IAMService) CreateUser(ctx context.Context, req *CreateUserReq) (*models.SysUser, error) {
	// 检查用户名唯一性
	existing, err := s.userDAO.FindByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if existing != nil {
		return nil, fmt.Errorf("[%d] 用户名已存在", errcode.ErrUserExists)
	}

	// 校验角色存在
	role, err := s.roleDAO.FindByID(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if role == nil {
		return nil, fmt.Errorf("[%d] 角色不存在", errcode.ErrInvalidParam)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	user := &models.SysUser{
		Username:     req.Username,
		PasswordHash: string(hash),
		RealName:     req.RealName,
		Email:        req.Email,
		Phone:        req.Phone,
		RoleID:       req.RoleID,
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	s.invalidateUserCache(ctx)

	return user, nil
}

func (s *IAMService) ListUsers(ctx context.Context, req *ListUserReq) (*PageResult, error) {
	p, ps, offset := response.NormalizePagination(req.Page, req.PageSize, 10, 100)
	req.Page, req.PageSize = p, ps

	// Normalize status: -1 means "all", convert to nil so DAO skips filter
	var statusVal int = -1
	var statusFilter *int
	if req.Status != nil {
		statusVal = *req.Status
	}
	if statusVal >= 0 {
		statusFilter = &statusVal
	}

	// Cache-aside: try Redis first
	cacheKey := fmt.Sprintf("user:list:%d:%d:%s:%d:%d", p, ps, req.Keyword, statusVal, req.RoleID)
	if s.redisClient != nil {
		if cached, err := s.redisClient.Get(ctx, cacheKey).Result(); err == nil {
			var result PageResult
			if json.Unmarshal([]byte(cached), &result) == nil {
				return &result, nil
			}
		}
	}

	users, total, err := s.userDAO.FindPage(offset, req.PageSize, req.Keyword, statusFilter, req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	dtos := make([]UserDTO, len(users))
	for i, u := range users {
		dtos[i] = toUserDTO(&u)
	}

	result := &PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     dtos,
	}

	// Write cache (TTL 60s)
	if s.redisClient != nil {
		if data, err := json.Marshal(result); err == nil {
			s.redisClient.Set(ctx, cacheKey, data, 60*time.Second)
		}
	}
	return result, nil
}

func (s *IAMService) GetUserByID(ctx context.Context, id uint, operatorID uint, operatorRole string) (*UserDTO, error) {
	user, err := s.userDAO.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if user == nil {
		return nil, fmt.Errorf("[%d] 用户不存在", errcode.ErrUserNotFound)
	}

	// 非管理员只能查看自己
	if operatorRole != "super_admin" && operatorRole != "lab_admin" && operatorID != id {
		return nil, fmt.Errorf("[%d] 权限不足", errcode.ErrPermissionDenied)
	}

	dto := toUserDTO(user)
	return &dto, nil
}

func (s *IAMService) UpdateUser(ctx context.Context, id uint, req *UpdateUserReq) error {
	user, err := s.userDAO.FindByID(id)
	if err != nil {
		return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if user == nil {
		return fmt.Errorf("[%d] 用户不存在", errcode.ErrUserNotFound)
	}

	updates := make(map[string]interface{})

	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.RoleID > 0 {
		role, err := s.roleDAO.FindByID(req.RoleID)
		if err != nil {
			return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
		}
		if role == nil {
			return fmt.Errorf("[%d] 角色不存在", errcode.ErrInvalidParam)
		}
		updates["role_id"] = req.RoleID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.userDAO.UpdateFields(id, updates); err != nil {
		return err
	}

	s.invalidateUserCache(ctx)
	return nil
}

func (s *IAMService) DisableUser(ctx context.Context, id uint, operatorID uint) error {
	if id == operatorID {
		return fmt.Errorf("[%d] 不能禁用自己", errcode.ErrInvalidParam)
	}

	user, err := s.userDAO.FindByID(id)
	if err != nil {
		return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if user == nil {
		return fmt.Errorf("[%d] 用户不存在", errcode.ErrUserNotFound)
	}

	// TODO V1.1: 检查是否有未归还借阅
	// TODO V1.1: 将该用户的活跃Token加入黑名单

	if err := s.userDAO.UpdateFields(id, map[string]interface{}{"status": 0}); err != nil {
		return err
	}

	s.invalidateUserCache(ctx)
	return nil
}

// invalidateUserCache 失效所有用户列表缓存（写操作后调用）
func (s *IAMService) invalidateUserCache(ctx context.Context) {
	if s.redisClient == nil {
		return
	}
	iter := s.redisClient.Scan(ctx, 0, "user:list:*", 100).Iterator()
	for iter.Next(ctx) {
		s.redisClient.Del(ctx, iter.Val())
	}
}

func (s *IAMService) ChangePassword(ctx context.Context, operatorID uint, targetID uint, req *ChangePasswordReq, isAdmin bool) error {
	if !isAdmin && operatorID != targetID {
		return fmt.Errorf("[%d] 权限不足", errcode.ErrPermissionDenied)
	}

	user, err := s.userDAO.FindByID(targetID)
	if err != nil {
		return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}
	if user == nil {
		return fmt.Errorf("[%d] 用户不存在", errcode.ErrUserNotFound)
	}

	// 非管理员必须验证旧密码
	if !isAdmin {
		if req.OldPassword == "" {
			return fmt.Errorf("[%d] 旧密码不能为空", errcode.ErrInvalidParam)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
			return fmt.Errorf("[%d] 旧密码错误", errcode.ErrInvalidParam)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("[%d] %w", errcode.ErrInternal, err)
	}

	return s.userDAO.UpdateFields(targetID, map[string]interface{}{"password_hash": string(hash)})
}

func toUserDTO(u *models.SysUser) UserDTO {
	dto := UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		RealName:  u.RealName,
		Email:     u.Email,
		Phone:     u.Phone,
		RoleID:    u.RoleID,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	if u.Role.ID != 0 {
		dto.Role = &RoleDTO{
			ID:          u.Role.ID,
			RoleName:    u.Role.RoleName,
			Description: u.Role.Description,
		}
	}
	return dto
}
