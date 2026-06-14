package dao

import (
	"errors"
	"fmt"

	"manage_system/models"
	"gorm.io/gorm"
)

type RoleDAO interface {
	FindAll() ([]models.SysRole, error)
	FindByID(id uint) (*models.SysRole, error)
	FindByName(name string) (*models.SysRole, error)
	Create(role *models.SysRole) error
	Update(role *models.SysRole) error
	Delete(id uint) error
}

type roleDAO struct {
	db *gorm.DB
}

func NewRoleDAO(db *gorm.DB) RoleDAO {
	return &roleDAO{db: db}
}

func (d *roleDAO) FindAll() ([]models.SysRole, error) {
	var roles []models.SysRole
	err := d.db.Order("id ASC").Find(&roles).Error
	return roles, err
}

func (d *roleDAO) FindByID(id uint) (*models.SysRole, error) {
	var role models.SysRole
	err := d.db.Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, err
}

func (d *roleDAO) FindByName(name string) (*models.SysRole, error) {
	var role models.SysRole
	err := d.db.Where("role_name = ?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, err
}

func (d *roleDAO) Create(role *models.SysRole) error {
	return d.db.Create(role).Error
}

func (d *roleDAO) Update(role *models.SysRole) error {
	// 系统预置角色：检查是否尝试修改 role_name
	if role.IsSystem == 1 {
		var old models.SysRole
		if err := d.db.Where("id = ?", role.ID).First(&old).Error; err != nil {
			return err
		}
		if old.RoleName != role.RoleName {
			return fmt.Errorf("不允许修改系统预置角色的 role_name")
		}
	}
	return d.db.Save(role).Error
}

func (d *roleDAO) Delete(id uint) error {
	role, err := d.FindByID(id)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("角色不存在")
	}
	if role.IsSystem == 1 {
		return fmt.Errorf("不允许删除系统预置角色")
	}

	// 检查是否有用户关联此角色，避免孤儿数据
	var count int64
	if err := d.db.Model(&models.SysUser{}).Where("role_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("查询角色关联用户失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("无法删除：该角色下仍有 %d 个用户，请先迁移用户", count)
	}

	return d.db.Delete(&models.SysRole{}, id).Error
}
