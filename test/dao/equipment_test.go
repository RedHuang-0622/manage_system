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

func setupEquipDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.LabEquipment{})
	require.NoError(t, err)
	return db
}

func TestEquipDAO_Insert_Success(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	equip := &models.LabEquipment{
		Name:           "GPU Server",
		Model:          "RTX4090",
		Category:       "服务器",
		TotalStock:     10,
		AvailableStock: 10,
		Location:       "A101",
		Status:         1,
	}

	err := d.Insert(equip)
	require.NoError(t, err)
	assert.NotZero(t, equip.ID)
}

func TestEquipDAO_Insert_ZeroStock(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	equip := &models.LabEquipment{
		Name:           "Empty Device",
		TotalStock:     0,
		AvailableStock: 0,
		Status:         1,
	}

	err := d.Insert(equip)
	require.NoError(t, err)
	assert.Equal(t, uint(0), equip.TotalStock)
}

func TestEquipDAO_FindByID_Exists(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "GPU", TotalStock: 5, AvailableStock: 5, Status: 1})

	equip, err := d.FindByID(1)
	require.NoError(t, err)
	require.NotNil(t, equip)
	assert.Equal(t, "GPU", equip.Name)
	assert.Equal(t, uint(5), equip.TotalStock)
}

func TestEquipDAO_FindByID_NotExists(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	equip, err := d.FindByID(9999)
	require.NoError(t, err)
	assert.Nil(t, equip)
}

func TestEquipDAO_FindPage_NoFilter(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	for i := 0; i < 5; i++ {
		d.Insert(&models.LabEquipment{Name: "Device", TotalStock: 5, AvailableStock: 5, Status: 1})
	}

	equips, total, err := d.FindPage(0, 12, "", "", nil, false)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, equips, 5)
}

func TestEquipDAO_FindPage_Keyword(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "GPU Server", Model: "A100", TotalStock: 5, AvailableStock: 5, Status: 1})
	d.Insert(&models.LabEquipment{Name: "CPU Server", Model: "Xeon", TotalStock: 3, AvailableStock: 3, Status: 1})
	d.Insert(&models.LabEquipment{Name: "Switch", Model: "S5700", TotalStock: 10, AvailableStock: 10, Status: 1})

	equips, total, err := d.FindPage(0, 10, "GPU", "", nil, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, equips, 1)
	assert.Equal(t, "GPU Server", equips[0].Name)
}

func TestEquipDAO_FindPage_Category(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "GPU1", Category: "服务器", TotalStock: 5, AvailableStock: 5, Status: 1})
	d.Insert(&models.LabEquipment{Name: "GPU2", Category: "服务器", TotalStock: 5, AvailableStock: 5, Status: 1})
	d.Insert(&models.LabEquipment{Name: "Switch1", Category: "网络设备", TotalStock: 10, AvailableStock: 10, Status: 1})

	equips, total, err := d.FindPage(0, 10, "", "服务器", nil, false)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, equips, 2)
}

func TestEquipDAO_FindPage_Status(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "Active", Status: 1, TotalStock: 5, AvailableStock: 5})
	d.Insert(&models.LabEquipment{Name: "Disabled", Status: 99, TotalStock: 5, AvailableStock: 5})
	d.UpdateFields(2, map[string]interface{}{"status": 0})

	status := 1
	equips, total, err := d.FindPage(0, 10, "", "", &status, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Active", equips[0].Name)
}

func TestEquipDAO_FindPage_OnlyAvailable(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "Available", TotalStock: 5, AvailableStock: 3, Status: 1})
	d.Insert(&models.LabEquipment{Name: "Depleted", TotalStock: 5, AvailableStock: 0, Status: 1})

	equips, total, err := d.FindPage(0, 10, "", "", nil, true)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Available", equips[0].Name)
}

func TestEquipDAO_FindPage_CombinedFilter(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "GPU Server A", Category: "服务器", TotalStock: 5, AvailableStock: 3, Status: 1})
	d.Insert(&models.LabEquipment{Name: "GPU Server B", Category: "服务器", TotalStock: 5, AvailableStock: 0, Status: 1})

	status := 1
	equips, total, err := d.FindPage(0, 10, "GPU", "服务器", &status, true)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "GPU Server A", equips[0].Name)
}

func TestEquipDAO_FindPage_EmptyResult(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	equips, total, err := d.FindPage(0, 10, "zzzz", "", nil, false)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, equips)
}

func TestEquipDAO_Update_Success(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "Old Name", TotalStock: 10, AvailableStock: 10, Status: 1})

	equip, _ := d.FindByID(1)
	equip.TotalStock = 20
	equip.AvailableStock = 15
	equip.Name = "New Name"
	err := d.Update(equip)
	require.NoError(t, err)

	updated, _ := d.FindByID(1)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, uint(20), updated.TotalStock)
	assert.Equal(t, uint(15), updated.AvailableStock)
}

func TestEquipDAO_UpdateFields_Partial(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "Device", Location: "A101", Status: 1, TotalStock: 5, AvailableStock: 5})

	err := d.UpdateFields(1, map[string]interface{}{"location": "B202"})
	require.NoError(t, err)

	updated, _ := d.FindByID(1)
	assert.Equal(t, "B202", updated.Location)
	assert.Equal(t, "Device", updated.Name) // unchanged
}

func TestEquipDAO_GetForUpdate_Exists(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	d.Insert(&models.LabEquipment{Name: "Device", TotalStock: 5, AvailableStock: 5, Status: 1})

	// SQLite doesn't support FOR UPDATE in the same way, but the clause is appended
	// For SQLite, this is effectively a normal read
	err := db.Transaction(func(tx *gorm.DB) error {
		equip, err := d.GetForUpdate(tx, 1)
		require.NoError(t, err)
		require.NotNil(t, equip)
		assert.Equal(t, "Device", equip.Name)
		return nil
	})
	require.NoError(t, err)
}

func TestEquipDAO_GetForUpdate_NotExists(t *testing.T) {
	db := setupEquipDB(t)
	d := dao.NewEquipmentDAO(db)

	err := db.Transaction(func(tx *gorm.DB) error {
		equip, err := d.GetForUpdate(tx, 9999)
		require.NoError(t, err)
		assert.Nil(t, equip)
		return nil
	})
	require.NoError(t, err)
}
