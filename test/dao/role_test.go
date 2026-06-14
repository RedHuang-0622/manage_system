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

func setupRoleDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.SysRole{}, &models.SysUser{})
	require.NoError(t, err)

	db.Create(&models.SysRole{ID: 1, RoleName: "super_admin", Description: "超级管理员", IsSystem: 1})
	db.Create(&models.SysRole{ID: 2, RoleName: "lab_admin", Description: "实验室负责人", IsSystem: 1})
	db.Create(&models.SysRole{ID: 3, RoleName: "member", Description: "普通成员", IsSystem: 1})

	return db
}

func TestRoleDAO_FindAll(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	roles, err := d.FindAll()
	require.NoError(t, err)
	assert.Len(t, roles, 3)
}

func TestRoleDAO_FindByID_Exists(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	role, err := d.FindByID(1)
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, "super_admin", role.RoleName)
}

func TestRoleDAO_FindByID_NotExists(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	role, err := d.FindByID(999)
	require.NoError(t, err)
	assert.Nil(t, role)
}

func TestRoleDAO_FindByName_Exists(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	role, err := d.FindByName("member")
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, uint(3), role.ID)
}

func TestRoleDAO_FindByName_NotExists(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	role, err := d.FindByName("nonexist_role")
	require.NoError(t, err)
	assert.Nil(t, role)
}

func TestRoleDAO_Create_Success(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	err := d.Create(&models.SysRole{RoleName: "custom_role", Description: "自定义角色"})
	require.NoError(t, err)

	roles, _ := d.FindAll()
	assert.Len(t, roles, 4)
}

func TestRoleDAO_Update_SystemRole_RejectsNameChange(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	role, _ := d.FindByID(1) // super_admin is system role
	role.RoleName = "hacked"
	err := d.Update(role)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不允许修改系统预置角色的 role_name")
}

func TestRoleDAO_Update_SystemRole_AllowsDescriptionChange(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	role, _ := d.FindByID(1)
	role.Description = "新的描述"
	role.RoleName = "super_admin" // keep same name
	err := d.Update(role)
	require.NoError(t, err)

	updated, _ := d.FindByID(1)
	assert.Equal(t, "新的描述", updated.Description)
}

func TestRoleDAO_Update_NonSystemRole(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	d.Create(&models.SysRole{RoleName: "custom", Description: "自定义", IsSystem: 0})

	role, _ := d.FindByID(4)
	role.RoleName = "renamed"
	err := d.Update(role)
	require.NoError(t, err)

	updated, _ := d.FindByID(4)
	assert.Equal(t, "renamed", updated.RoleName)
}

func TestRoleDAO_Delete_SystemRole_Rejected(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	err := d.Delete(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不允许删除系统预置角色")
}

func TestRoleDAO_Delete_NonSystemRole_Success(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	d.Create(&models.SysRole{RoleName: "to_delete", IsSystem: 0})

	err := d.Delete(4)
	require.NoError(t, err)

	role, _ := d.FindByID(4)
	assert.Nil(t, role)
}

func TestRoleDAO_Delete_NotExists(t *testing.T) {
	db := setupRoleDB(t)
	d := dao.NewRoleDAO(db)

	err := d.Delete(999)
	assert.Error(t, err)
}
