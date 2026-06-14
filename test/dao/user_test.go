package dao_test

import (
	"testing"

	"manage_system/dao"
	"manage_system/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupUserDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.SysUser{}, &models.SysRole{})
	require.NoError(t, err)

	// Seed roles
	db.Create(&models.SysRole{ID: 1, RoleName: "super_admin", Description: "超级管理员"})
	db.Create(&models.SysRole{ID: 2, RoleName: "lab_admin", Description: "实验室负责人"})
	db.Create(&models.SysRole{ID: 3, RoleName: "member", Description: "普通成员"})

	return db
}

func TestUserDAO_Create_Success(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	user := &models.SysUser{
		Username: "zhangsan",
		RealName: "张三",
		RoleID:   3,
		Status:   1,
	}

	err := d.Create(user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserDAO_Create_DuplicateUsername(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	u1 := &models.SysUser{Username: "admin", RoleID: 1, Status: 1}
	err := d.Create(u1)
	require.NoError(t, err)

	u2 := &models.SysUser{Username: "admin", RoleID: 2, Status: 1}
	err = d.Create(u2)
	assert.Error(t, err, "duplicate username should fail")
}

func TestUserDAO_FindByUsername_Exists(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	d.Create(&models.SysUser{Username: "zhangsan", RealName: "张三", RoleID: 3, Status: 1})

	user, err := d.FindByUsername("zhangsan")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "zhangsan", user.Username)
	assert.Equal(t, "张三", user.RealName)
}

func TestUserDAO_FindByUsername_NotExists(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	user, err := d.FindByUsername("nonexist")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserDAO_FindByID_Exists(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	d.Create(&models.SysUser{Username: "testuser", RealName: "测试", RoleID: 3, Status: 1})

	user, err := d.FindByID(1)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.NotZero(t, user.Role.ID, "should preload Role")
}

func TestUserDAO_FindByID_NotExists(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	user, err := d.FindByID(9999)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserDAO_FindPage_NoFilter(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	for i := 0; i < 5; i++ {
		d.Create(&models.SysUser{Username: "user" + string(rune('a'+i)), RoleID: 3, Status: 1})
	}

	users, total, err := d.FindPage(0, 10, "", nil, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 5)
}

func TestUserDAO_FindPage_Keyword(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	d.Create(&models.SysUser{Username: "zhangsan", RealName: "张三", RoleID: 3, Status: 1})
	d.Create(&models.SysUser{Username: "lisi", RealName: "李四", RoleID: 3, Status: 1})
	d.Create(&models.SysUser{Username: "wangwu", RealName: "王五", RoleID: 3, Status: 1})

	users, total, err := d.FindPage(0, 10, "张", nil, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, "zhangsan", users[0].Username)
}

func TestUserDAO_FindPage_StatusFilter(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	d.Create(&models.SysUser{Username: "active1", RoleID: 3, Status: 1})
	d.Create(&models.SysUser{Username: "active2", RoleID: 3, Status: 1})
	// GORM uses default value (1) for zero-value fields like Status:0
	// Use UpdateFields to set status to 0 explicitly
	d.Create(&models.SysUser{Username: "disabled1", RoleID: 3, Status: 99}) // temporary placeholder
	d.UpdateFields(3, map[string]interface{}{"status": 0})

	status := 1
	users, total, err := d.FindPage(0, 10, "", &status, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
}

func TestUserDAO_FindPage_RoleFilter(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	d.Create(&models.SysUser{Username: "admin2", RoleID: 1, Status: 1})
	d.Create(&models.SysUser{Username: "member1", RoleID: 3, Status: 1})
	d.Create(&models.SysUser{Username: "member2", RoleID: 3, Status: 1})

	users, total, err := d.FindPage(0, 10, "", nil, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
}

func TestUserDAO_FindPage_EmptyResult(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	users, total, err := d.FindPage(0, 10, "zzzznotfound", nil, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, users)
}

func TestUserDAO_FindPage_Pagination(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	for i := 0; i < 5; i++ {
		d.Create(&models.SysUser{Username: "user" + string(rune('a'+i)), RoleID: 3, Status: 1})
	}

	users, total, err := d.FindPage(0, 2, "", nil, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 2)
}

func TestUserDAO_UpdateFields_Success(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	d.Create(&models.SysUser{Username: "testuser", Email: "old@test.com", RoleID: 3, Status: 1})

	err := d.UpdateFields(1, map[string]interface{}{"email": "new@test.com"})
	require.NoError(t, err)

	user, _ := d.FindByID(1)
	assert.Equal(t, "new@test.com", user.Email)
}

func TestUserDAO_UpdateFields_Status(t *testing.T) {
	db := setupUserDB(t)
	d := dao.NewUserDAO(db)

	d.Create(&models.SysUser{Username: "testuser", RoleID: 3, Status: 1})

	err := d.UpdateFields(1, map[string]interface{}{"status": 0})
	require.NoError(t, err)

	user, _ := d.FindByID(1)
	assert.Equal(t, int8(0), user.Status)
}
