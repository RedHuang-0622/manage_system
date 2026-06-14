package models

import "time"

type SysUser struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	RealName     string    `gorm:"size:32" json:"real_name"`
	Email        string    `gorm:"size:64" json:"email"`
	Phone        string    `gorm:"size:16" json:"phone"`
	RoleID       uint      `gorm:"not null;default:0;index" json:"role_id"`
	Status       int8      `gorm:"not null;default:1;index" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Role         SysRole   `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

type SysRole struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleName    string    `gorm:"size:32;uniqueIndex;not null" json:"role_name"`
	Description string    `gorm:"size:128" json:"description"`
	IsSystem    int8      `gorm:"not null;default:0" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
