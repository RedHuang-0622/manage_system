package dao

import (
	"errors"

	"manage_system/models"
	"gorm.io/gorm"
)

type UserDAO interface {
	Create(user *models.SysUser) error
	FindByID(id uint) (*models.SysUser, error)
	FindByUsername(username string) (*models.SysUser, error)
	FindPage(offset, limit int, keyword string, status *int, roleID uint) ([]models.SysUser, int64, error)
	UpdateFields(id uint, updates map[string]interface{}) error
	Update(user *models.SysUser) error
}

type userDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &userDAO{db: db}
}

func (d *userDAO) Create(user *models.SysUser) error {
	return d.db.Create(user).Error
}

func (d *userDAO) FindByID(id uint) (*models.SysUser, error) {
	var user models.SysUser
	err := d.db.Preload("Role").Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (d *userDAO) FindByUsername(username string) (*models.SysUser, error) {
	var user models.SysUser
	err := d.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (d *userDAO) FindPage(offset, limit int, keyword string, status *int, roleID uint) ([]models.SysUser, int64, error) {
	query := d.db.Model(&models.SysUser{}).Preload("Role")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR real_name LIKE ?", like, like)
	}
	if status != nil && *status >= 0 {
		query = query.Where("status = ?", *status)
	}
	if roleID > 0 {
		query = query.Where("role_id = ?", roleID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []models.SysUser
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (d *userDAO) UpdateFields(id uint, updates map[string]interface{}) error {
	return d.db.Model(&models.SysUser{}).Where("id = ?", id).Updates(updates).Error
}

func (d *userDAO) Update(user *models.SysUser) error {
	return d.db.Save(user).Error
}
