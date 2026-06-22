package dao

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"manage_system/models"
)

type EquipmentDAO interface {
	Insert(equip *models.LabEquipment) error
	FindByID(id uint) (*models.LabEquipment, error)
	FindPage(offset, limit int, keyword, category string, status *int, onlyAvailable bool) ([]models.LabEquipment, int64, error)
	Update(equip *models.LabEquipment) error
	UpdateFields(id uint, updates map[string]interface{}) error
	GetForUpdate(tx *gorm.DB, id uint) (*models.LabEquipment, error)
	UpdateWithTx(tx *gorm.DB, equip *models.LabEquipment) error
}

type equipmentDAO struct {
	db *gorm.DB
}

func NewEquipmentDAO(db *gorm.DB) EquipmentDAO {
	return &equipmentDAO{db: db}
}

func (d *equipmentDAO) Insert(equip *models.LabEquipment) error {
	return d.db.Create(equip).Error
}

func (d *equipmentDAO) FindByID(id uint) (*models.LabEquipment, error) {
	var equip models.LabEquipment
	err := d.db.Where("id = ?", id).First(&equip).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &equip, err
}

func (d *equipmentDAO) FindPage(offset, limit int, keyword, category string, status *int, onlyAvailable bool) ([]models.LabEquipment, int64, error) {
	query := d.db.Model(&models.LabEquipment{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR model LIKE ?", like, like)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != nil && *status >= 0 {
		query = query.Where("status = ?", *status)
	}
	if onlyAvailable {
		query = query.Where("available_stock > 0")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var equips []models.LabEquipment
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&equips).Error
	return equips, total, err
}

func (d *equipmentDAO) Update(equip *models.LabEquipment) error {
	return d.db.Save(equip).Error
}

func (d *equipmentDAO) UpdateFields(id uint, updates map[string]interface{}) error {
	return d.db.Model(&models.LabEquipment{}).Where("id = ?", id).Updates(updates).Error
}

func (d *equipmentDAO) GetForUpdate(tx *gorm.DB, id uint) (*models.LabEquipment, error) {
	var equip models.LabEquipment
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&equip).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &equip, err
}

func (d *equipmentDAO) UpdateWithTx(tx *gorm.DB, equip *models.LabEquipment) error {
	return tx.Save(equip).Error
}
