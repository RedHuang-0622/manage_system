package dao_test

import (
	"testing"
	"time"

	"manage_system/dao"
	"manage_system/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupBorrowDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.BorrowRecord{}, &models.SysUser{}, &models.SysRole{}, &models.LabEquipment{})
	require.NoError(t, err)

	// Seed data
	db.Create(&models.SysUser{ID: 1, Username: "admin", RoleID: 1, Status: 1})
	db.Create(&models.SysUser{ID: 2, Username: "member1", RoleID: 3, Status: 1})
	db.Create(&models.LabEquipment{ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 10, Status: 1})
	db.Create(&models.LabEquipment{ID: 2, Name: "CPU", TotalStock: 5, AvailableStock: 5, Status: 1})

	return db
}

func TestBorrowDAO_Insert_Success(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	record := &models.BorrowRecord{
		UserID:      2,
		EquipmentID: 1,
		Quantity:    2,
		Status:      models.BorrowStatusApplied,
		ApplyNote:   "test borrow",
		ApplyAt:     time.Now(),
	}

	err := d.Insert(record)
	require.NoError(t, err)
	assert.NotZero(t, record.ID)
	assert.Equal(t, models.BorrowStatusApplied, record.Status)
}

func TestBorrowDAO_FindByID_Exists(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{
		UserID: 2, EquipmentID: 1, Quantity: 1,
		Status: models.BorrowStatusApplied, ApplyAt: time.Now(),
	})

	record, err := d.FindByID(1)
	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, uint(2), record.UserID)
	assert.NotZero(t, record.User.ID, "should preload User")
	assert.NotZero(t, record.Equipment.ID, "should preload Equipment")
}

func TestBorrowDAO_FindByID_NotExists(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	record, err := d.FindByID(9999)
	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestBorrowDAO_FindPage_ByUserID(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 1, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})
	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})
	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 2, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})

	records, total, err := d.FindPage(0, 10, 2, 0, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, records, 2)
	for _, r := range records {
		assert.Equal(t, uint(2), r.UserID)
	}
}

func TestBorrowDAO_FindPage_ByEquipmentID(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 1, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})
	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})
	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 2, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})

	records, total, err := d.FindPage(0, 10, 0, 1, "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, records, 2)
}

func TestBorrowDAO_FindPage_ByStatus(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})
	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusBorrowed, ApplyAt: time.Now()})

	records, total, err := d.FindPage(0, 10, 0, 0, "已借出")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, models.BorrowStatusBorrowed, records[0].Status)
}

func TestBorrowDAO_FindPage_CombinedFilter(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusBorrowed, ApplyAt: time.Now()})
	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})

	records, total, err := d.FindPage(0, 10, 2, 1, "已借出")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, models.BorrowStatusBorrowed, records[0].Status)
}

func TestBorrowDAO_CountActiveByUserAndEquipment_HasActive(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})

	count, err := d.CountActiveByUserAndEquipment(2, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestBorrowDAO_CountActiveByUserAndEquipment_NoActive(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusReturned, ApplyAt: time.Now()})

	count, err := d.CountActiveByUserAndEquipment(2, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestBorrowDAO_CountByEquipmentAndStatus(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusBorrowed, ApplyAt: time.Now()})
	d.Insert(&models.BorrowRecord{UserID: 1, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusBorrowed, ApplyAt: time.Now()})

	count, err := d.CountByEquipmentAndStatus(1, "已借出")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestBorrowDAO_CountByEquipmentAndStatus_Zero(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	count, err := d.CountByEquipmentAndStatus(1, "已归还")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestBorrowDAO_GetForUpdate_Exists(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})

	err := db.Transaction(func(tx *gorm.DB) error {
		record, err := d.GetForUpdate(tx, 1)
		require.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, uint(2), record.UserID)
		return nil
	})
	require.NoError(t, err)
}

func TestBorrowDAO_GetForUpdate_NotExists(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	err := db.Transaction(func(tx *gorm.DB) error {
		record, err := d.GetForUpdate(tx, 9999)
		require.NoError(t, err)
		assert.Nil(t, record)
		return nil
	})
	require.NoError(t, err)
}

func TestBorrowDAO_UpdateWithTx_Success(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusApplied, ApplyAt: time.Now()})

	err := db.Transaction(func(tx *gorm.DB) error {
		record, _ := d.GetForUpdate(tx, 1)
		record.Status = models.BorrowStatusBorrowed
		return d.UpdateWithTx(tx, record)
	})
	require.NoError(t, err)

	record, _ := d.FindByID(1)
	assert.Equal(t, models.BorrowStatusBorrowed, record.Status)
}

func TestBorrowDAO_CountByEquipmentAndStatusInTx(t *testing.T) {
	db := setupBorrowDB(t)
	d := dao.NewBorrowDAO(db)

	d.Insert(&models.BorrowRecord{UserID: 2, EquipmentID: 1, Quantity: 1, Status: models.BorrowStatusBorrowed, ApplyAt: time.Now()})

	err := db.Transaction(func(tx *gorm.DB) error {
		count, err := d.CountByEquipmentAndStatusInTx(tx, 1, "已借出")
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		return nil
	})
	require.NoError(t, err)
}
