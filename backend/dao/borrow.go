package dao

import (
	"errors"

	"manage_system/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BorrowDAO interface {
	Insert(record *models.BorrowRecord) error
	FindByID(id uint) (*models.BorrowRecord, error)
	FindPage(offset, limit int, userID, equipID uint, status string) ([]models.BorrowRecord, int64, error)
	CountActiveByUserAndEquipment(userID, equipID uint) (int64, error)
	CountByEquipmentAndStatus(equipID uint, status string) (int64, error)
	CountByEquipmentAndStatusInTx(tx *gorm.DB, equipID uint, status string) (int64, error)
	GetForUpdate(tx *gorm.DB, id uint) (*models.BorrowRecord, error)
	UpdateWithTx(tx *gorm.DB, record *models.BorrowRecord) error
}

type borrowDAO struct {
	db *gorm.DB
}

func NewBorrowDAO(db *gorm.DB) BorrowDAO {
	return &borrowDAO{db: db}
}

func (d *borrowDAO) Insert(record *models.BorrowRecord) error {
	record.ApplyAt = record.ApplyAt.UTC()
	return d.db.Create(record).Error
}

func (d *borrowDAO) FindByID(id uint) (*models.BorrowRecord, error) {
	var record models.BorrowRecord
	err := d.db.Preload("User").Preload("Equipment").Preload("Approver").
		Where("id = ?", id).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

func (d *borrowDAO) FindPage(offset, limit int, userID, equipID uint, status string) ([]models.BorrowRecord, int64, error) {
	query := d.db.Model(&models.BorrowRecord{}).Preload("User").Preload("Equipment")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if equipID > 0 {
		query = query.Where("equipment_id = ?", equipID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []models.BorrowRecord
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&records).Error
	return records, total, err
}

func (d *borrowDAO) CountActiveByUserAndEquipment(userID, equipID uint) (int64, error) {
	var count int64
	err := d.db.Model(&models.BorrowRecord{}).
		Where("user_id = ? AND equipment_id = ? AND status IN ?",
			userID, equipID, []string{"申请中", "已借出"}).
		Count(&count).Error
	return count, err
}

func (d *borrowDAO) CountByEquipmentAndStatus(equipID uint, status string) (int64, error) {
	var count int64
	err := d.db.Model(&models.BorrowRecord{}).
		Where("equipment_id = ? AND status = ?", equipID, status).
		Count(&count).Error
	return count, err
}

func (d *borrowDAO) CountByEquipmentAndStatusInTx(tx *gorm.DB, equipID uint, status string) (int64, error) {
	var count int64
	err := tx.Model(&models.BorrowRecord{}).
		Where("equipment_id = ? AND status = ?", equipID, status).
		Count(&count).Error
	return count, err
}

func (d *borrowDAO) GetForUpdate(tx *gorm.DB, id uint) (*models.BorrowRecord, error) {
	var record models.BorrowRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

func (d *borrowDAO) UpdateWithTx(tx *gorm.DB, record *models.BorrowRecord) error {
	return tx.Save(record).Error
}
