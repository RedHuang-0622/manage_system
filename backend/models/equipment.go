package models

import "time"

type LabEquipment struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"size:128;not null;index" json:"name"`
	Model          string    `gorm:"size:64" json:"model"`
	Category       string    `gorm:"size:32;index" json:"category"`
	TotalStock     uint      `gorm:"not null;default:0" json:"total_stock"`
	AvailableStock uint      `gorm:"not null;default:0" json:"available_stock"`
	Location       string    `gorm:"size:64" json:"location"`
	Description    string    `gorm:"type:text" json:"description"`
	Status         int8      `gorm:"not null;default:1" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
