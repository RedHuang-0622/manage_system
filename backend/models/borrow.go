package models

import "time"

type BorrowStatus string

const (
	BorrowStatusApplied  BorrowStatus = "申请中"
	BorrowStatusBorrowed BorrowStatus = "已借出"
	BorrowStatusReturned BorrowStatus = "已归还"
	BorrowStatusRejected BorrowStatus = "被拒绝"
	BorrowStatusCanceled BorrowStatus = "已取消"
)

// ValidTransitions 合法的状态转换
var ValidTransitions = map[BorrowStatus][]BorrowStatus{
	BorrowStatusApplied:  {BorrowStatusBorrowed, BorrowStatusRejected, BorrowStatusCanceled},
	BorrowStatusBorrowed: {BorrowStatusReturned},
}

// IsTerminal 是否为终态
func (s BorrowStatus) IsTerminal() bool {
	return s == BorrowStatusReturned || s == BorrowStatusRejected || s == BorrowStatusCanceled
}

type BorrowRecord struct {
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint         `gorm:"not null;index" json:"user_id"`
	EquipmentID uint         `gorm:"not null;index" json:"equipment_id"`
	Quantity    uint         `gorm:"not null;default:1" json:"quantity"`
	Status      BorrowStatus `gorm:"size:32;not null;default:'申请中';index" json:"status"`
	ApplyNote   string       `gorm:"size:256" json:"apply_note"`
	ApproveNote string       `gorm:"size:256" json:"approve_note"`
	ApproverID  *uint        `gorm:"default:null" json:"approver_id"`
	ApplyAt     time.Time    `gorm:"not null;index" json:"apply_at"`
	ApproveAt   *time.Time   `json:"approve_at"`
	ReturnAt    *time.Time   `json:"return_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`

	User      SysUser      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Equipment LabEquipment `gorm:"foreignKey:EquipmentID" json:"equipment,omitempty"`
	Approver  SysUser      `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}
