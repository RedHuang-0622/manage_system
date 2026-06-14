package service_test

import (
	"context"
	"testing"

	"manage_system/models"
	"manage_system/pkg/config"
	jwtpkg "manage_system/pkg/jwt"
	"manage_system/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ──────────────────────── Mock DAOs ────────────────────────

type MockUserDAO struct{ mock.Mock }

func (m *MockUserDAO) Create(user *models.SysUser) error {
	args := m.Called(user)
	if args.Error(0) == nil {
		user.ID = 1
	}
	return args.Error(0)
}
func (m *MockUserDAO) FindByID(id uint) (*models.SysUser, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SysUser), args.Error(1)
}
func (m *MockUserDAO) FindByUsername(username string) (*models.SysUser, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SysUser), args.Error(1)
}
func (m *MockUserDAO) FindPage(offset, limit int, keyword string, status *int, roleID uint) ([]models.SysUser, int64, error) {
	args := m.Called(offset, limit, keyword, status, roleID)
	return args.Get(0).([]models.SysUser), args.Get(1).(int64), args.Error(2)
}
func (m *MockUserDAO) UpdateFields(id uint, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}
func (m *MockUserDAO) Update(user *models.SysUser) error {
	args := m.Called(user)
	return args.Error(0)
}

type MockRoleDAO struct{ mock.Mock }

func (m *MockRoleDAO) FindAll() ([]models.SysRole, error) {
	args := m.Called()
	return args.Get(0).([]models.SysRole), args.Error(1)
}
func (m *MockRoleDAO) FindByID(id uint) (*models.SysRole, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SysRole), args.Error(1)
}
func (m *MockRoleDAO) FindByName(name string) (*models.SysRole, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SysRole), args.Error(1)
}
func (m *MockRoleDAO) Create(role *models.SysRole) error {
	args := m.Called(role)
	return args.Error(0)
}
func (m *MockRoleDAO) Update(role *models.SysRole) error {
	args := m.Called(role)
	return args.Error(0)
}
func (m *MockRoleDAO) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupIAMService(t *testing.T) (*service.IAMService, *MockUserDAO, *MockRoleDAO, *miniredis.Miniredis) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	userDAO := new(MockUserDAO)
	roleDAO := new(MockRoleDAO)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := config.JWTConfig{
		Secret: "test-secret-key-at-least-32-chars-long!!",
		Expire: 7200,
		Issuer: "test-issuer",
	}
	jwtService := jwtpkg.NewService(cfg, rdb)

	svc := service.NewIAMService(db, userDAO, roleDAO, jwtService, rdb)
	return svc, userDAO, roleDAO, mr
}

// ──────────────────────── Login Tests ────────────────────────

func TestIAMService_Login_Success(t *testing.T) {
	svc, userDAO, roleDAO, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
	user := &models.SysUser{ID: 1, Username: "admin", PasswordHash: string(hash), RoleID: 1, Status: 1}
	role := &models.SysRole{ID: 1, RoleName: "super_admin"}

	userDAO.On("FindByUsername", "admin").Return(user, nil)
	roleDAO.On("FindByID", uint(1)).Return(role, nil)

	resp, err := svc.Login(context.Background(), &service.LoginReq{
		Username: "admin",
		Password: "admin123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Greater(t, resp.ExpiresIn, int64(0))
}

func TestIAMService_Login_UserNotFound(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	userDAO.On("FindByUsername", "nobody").Return(nil, nil)

	resp, err := svc.Login(context.Background(), &service.LoginReq{
		Username: "nobody",
		Password: "password",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户名或密码错误")
}

func TestIAMService_Login_WrongPassword(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), 12)
	user := &models.SysUser{ID: 1, Username: "admin", PasswordHash: string(hash), RoleID: 1, Status: 1}
	userDAO.On("FindByUsername", "admin").Return(user, nil)

	resp, err := svc.Login(context.Background(), &service.LoginReq{
		Username: "admin",
		Password: "wrong_password",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户名或密码错误")
}

func TestIAMService_Login_AccountDisabled(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
	user := &models.SysUser{ID: 1, Username: "admin", PasswordHash: string(hash), RoleID: 1, Status: 0}
	userDAO.On("FindByUsername", "admin").Return(user, nil)

	resp, err := svc.Login(context.Background(), &service.LoginReq{
		Username: "admin",
		Password: "admin123",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "账号已被禁用")
}

// ──────────────────────── Logout Tests ────────────────────────

func TestIAMService_Logout_Success(t *testing.T) {
	svc, _, _, _ := setupIAMService(t)

	// Logout with an invalid token should not error (graceful handling)
	err := svc.Logout(context.Background(), "some-invalid-token-string")
	assert.NoError(t, err, "logout should not error even with invalid token")
}

// ──────────────────────── Refresh Token Tests ────────────────────────

func TestIAMService_RefreshToken_InvalidToken(t *testing.T) {
	svc, _, _, _ := setupIAMService(t)

	resp, err := svc.RefreshToken(context.Background(), "invalid-token")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token无效或已过期")
}

// ──────────────────────── ListRoles Tests ────────────────────────

func TestIAMService_ListRoles_Success(t *testing.T) {
	svc, _, roleDAO, _ := setupIAMService(t)

	roles := []models.SysRole{
		{ID: 1, RoleName: "super_admin", Description: "超级管理员"},
		{ID: 2, RoleName: "lab_admin", Description: "实验室负责人"},
		{ID: 3, RoleName: "member", Description: "普通成员"},
	}
	roleDAO.On("FindAll").Return(roles, nil)

	result, err := svc.ListRoles(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "super_admin", result[0].RoleName)
}

// ──────────────────────── CreateUser Tests ────────────────────────

func TestIAMService_CreateUser_Success(t *testing.T) {
	svc, userDAO, roleDAO, _ := setupIAMService(t)

	userDAO.On("FindByUsername", "newuser").Return(nil, nil)
	roleDAO.On("FindByID", uint(3)).Return(&models.SysRole{ID: 3, RoleName: "member"}, nil)
	userDAO.On("Create", mock.AnythingOfType("*models.SysUser")).Return(nil)

	user, err := svc.CreateUser(context.Background(), &service.CreateUserReq{
		Username: "newuser",
		Password: "password123",
		RealName: "新用户",
		RoleID:   3,
	})

	require.NoError(t, err)
	assert.Equal(t, "newuser", user.Username)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "password123", user.PasswordHash, "password should be hashed")
}

func TestIAMService_CreateUser_UsernameExists(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), 12)
	existing := &models.SysUser{ID: 1, Username: "admin", PasswordHash: string(hash), RoleID: 1, Status: 1}
	userDAO.On("FindByUsername", "admin").Return(existing, nil)

	user, err := svc.CreateUser(context.Background(), &service.CreateUserReq{
		Username: "admin",
		Password: "password123",
		RealName: "重复用户",
		RoleID:   3,
	})

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户名已存在")
}

func TestIAMService_CreateUser_RoleNotFound(t *testing.T) {
	svc, userDAO, roleDAO, _ := setupIAMService(t)

	userDAO.On("FindByUsername", "newuser").Return(nil, nil)
	roleDAO.On("FindByID", uint(999)).Return(nil, nil)

	user, err := svc.CreateUser(context.Background(), &service.CreateUserReq{
		Username: "newuser",
		Password: "password123",
		RealName: "新用户",
		RoleID:   999,
	})

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "角色不存在")
}

func TestIAMService_CreateUser_PasswordHashed(t *testing.T) {
	svc, userDAO, roleDAO, _ := setupIAMService(t)

	userDAO.On("FindByUsername", "user").Return(nil, nil)
	roleDAO.On("FindByID", uint(3)).Return(&models.SysRole{ID: 3, RoleName: "member"}, nil)
	userDAO.On("Create", mock.AnythingOfType("*models.SysUser")).Return(nil)

	user, err := svc.CreateUser(context.Background(), &service.CreateUserReq{
		Username: "user",
		Password: "secret123",
		RealName: "User",
		RoleID:   3,
	})

	require.NoError(t, err)
	// password_hash should be bcrypt hash, not plaintext
	assert.NotEqual(t, "secret123", user.PasswordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret123"))
	assert.NoError(t, err, "should be valid bcrypt hash")
}

// ──────────────────────── ListUsers Tests ────────────────────────

func TestIAMService_ListUsers_Success(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	users := []models.SysUser{
		{ID: 1, Username: "user1", RealName: "用户1", RoleID: 3, Status: 1},
		{ID: 2, Username: "user2", RealName: "用户2", RoleID: 3, Status: 1},
	}
	userDAO.On("FindPage", 0, 10, "", (*int)(nil), uint(0)).Return(users, int64(2), nil)

	result, err := svc.ListUsers(context.Background(), &service.ListUserReq{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.List, 2)
}

// ──────────────────────── GetUserByID Tests ────────────────────────

func TestIAMService_GetUserByID_Success(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	user := &models.SysUser{ID: 1, Username: "test", RealName: "测试", RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(1)).Return(user, nil)

	dto, err := svc.GetUserByID(context.Background(), 1, 1, "super_admin")
	require.NoError(t, err)
	assert.Equal(t, "test", dto.Username)
}

func TestIAMService_GetUserByID_NotFound(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	userDAO.On("FindByID", uint(9999)).Return(nil, nil)

	dto, err := svc.GetUserByID(context.Background(), 9999, 1, "super_admin")
	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
}

func TestIAMService_GetUserByID_PermissionDenied(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	user := &models.SysUser{ID: 2, Username: "other_user", RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(2)).Return(user, nil)

	// member trying to view another user's info
	dto, err := svc.GetUserByID(context.Background(), 2, 3, "member")
	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "权限不足")
}

// ──────────────────────── UpdateUser Tests ────────────────────────

func TestIAMService_UpdateUser_Success(t *testing.T) {
	svc, userDAO, roleDAO, _ := setupIAMService(t)

	user := &models.SysUser{ID: 1, Username: "test", RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(1)).Return(user, nil)
	roleDAO.On("FindByID", uint(2)).Return(&models.SysRole{ID: 2, RoleName: "lab_admin"}, nil)
	userDAO.On("UpdateFields", uint(1), mock.AnythingOfType("map[string]interface {}")).Return(nil)

	err := svc.UpdateUser(context.Background(), 1, &service.UpdateUserReq{
		RealName: "新名字",
		RoleID:   2,
	})
	require.NoError(t, err)
}

func TestIAMService_UpdateUser_NotFound(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	userDAO.On("FindByID", uint(9999)).Return(nil, nil)

	err := svc.UpdateUser(context.Background(), 9999, &service.UpdateUserReq{
		RealName: "xxx",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
}

// ──────────────────────── DisableUser Tests ────────────────────────

func TestIAMService_DisableUser_Success(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	user := &models.SysUser{ID: 3, Username: "member1", RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(3)).Return(user, nil)
	userDAO.On("UpdateFields", uint(3), mock.AnythingOfType("map[string]interface {}")).Return(nil)

	err := svc.DisableUser(context.Background(), 3, 1)
	require.NoError(t, err)
}

func TestIAMService_DisableUser_SelfDisable(t *testing.T) {
	svc, _, _, _ := setupIAMService(t)

	err := svc.DisableUser(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不能禁用自己")
}

func TestIAMService_DisableUser_NotFound(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	userDAO.On("FindByID", uint(9999)).Return(nil, nil)

	err := svc.DisableUser(context.Background(), 9999, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
}

// ──────────────────────── ChangePassword Tests ────────────────────────

func TestIAMService_ChangePassword_Success(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), 12)
	user := &models.SysUser{ID: 3, Username: "member1", PasswordHash: string(hash), RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(3)).Return(user, nil)
	userDAO.On("UpdateFields", uint(3), mock.AnythingOfType("map[string]interface {}")).Return(nil)

	err := svc.ChangePassword(context.Background(), 3, 3, &service.ChangePasswordReq{
		OldPassword: "oldpass",
		NewPassword: "newpass123",
	}, false)
	require.NoError(t, err)
}

func TestIAMService_ChangePassword_WrongOldPassword(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctold"), 12)
	user := &models.SysUser{ID: 3, Username: "member1", PasswordHash: string(hash), RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(3)).Return(user, nil)

	err := svc.ChangePassword(context.Background(), 3, 3, &service.ChangePasswordReq{
		OldPassword: "wrongold",
		NewPassword: "newpass123",
	}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "旧密码错误")
}

func TestIAMService_ChangePassword_AdminSkipOldPassword(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("anyold"), 12)
	user := &models.SysUser{ID: 3, Username: "member1", PasswordHash: string(hash), RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(3)).Return(user, nil)
	userDAO.On("UpdateFields", uint(3), mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// Admin doesn't need old password
	err := svc.ChangePassword(context.Background(), 1, 3, &service.ChangePasswordReq{
		NewPassword: "newpass123",
	}, true)
	require.NoError(t, err)
}

func TestIAMService_ChangePassword_NonAdminMissingOld(t *testing.T) {
	svc, userDAO, _, _ := setupIAMService(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("anyold"), 12)
	user := &models.SysUser{ID: 3, Username: "member1", PasswordHash: string(hash), RoleID: 3, Status: 1}
	userDAO.On("FindByID", uint(3)).Return(user, nil)

	err := svc.ChangePassword(context.Background(), 3, 3, &service.ChangePasswordReq{
		NewPassword: "newpass123",
	}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "旧密码不能为空")
}

func TestIAMService_ChangePassword_PermissionDenied(t *testing.T) {
	svc, _, _, _ := setupIAMService(t)

	// Non-admin trying to change other user's password
	err := svc.ChangePassword(context.Background(), 3, 2, &service.ChangePasswordReq{
		OldPassword: "old",
		NewPassword: "newpass123",
	}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "权限不足")
}
