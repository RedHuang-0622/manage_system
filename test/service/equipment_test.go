package service_test

import (
	"context"
	"testing"

	"manage_system/models"
	"manage_system/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ──────────────────────── Mock EquipmentDAO ────────────────────────

type MockEquipDAO struct{ mock.Mock }

func (m *MockEquipDAO) Insert(equip *models.LabEquipment) error {
	args := m.Called(equip)
	if args.Error(0) == nil {
		equip.ID = 1
	}
	return args.Error(0)
}
func (m *MockEquipDAO) FindByID(id uint) (*models.LabEquipment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LabEquipment), args.Error(1)
}
func (m *MockEquipDAO) FindPage(offset, limit int, keyword, category string, status *int, onlyAvailable bool) ([]models.LabEquipment, int64, error) {
	args := m.Called(offset, limit, keyword, category, status, onlyAvailable)
	return args.Get(0).([]models.LabEquipment), args.Get(1).(int64), args.Error(2)
}
func (m *MockEquipDAO) Update(equip *models.LabEquipment) error {
	args := m.Called(equip)
	return args.Error(0)
}
func (m *MockEquipDAO) UpdateFields(id uint, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}
func (m *MockEquipDAO) GetForUpdate(tx *gorm.DB, id uint) (*models.LabEquipment, error) {
	args := m.Called(tx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LabEquipment), args.Error(1)
}
func (m *MockEquipDAO) UpdateWithTx(tx *gorm.DB, equip *models.LabEquipment) error {
	args := m.Called(tx, equip)
	return args.Error(0)
}

// ──────────────────────── Mock BorrowStatusChecker ────────────────────────

type MockBorrowChecker struct{ mock.Mock }

func (m *MockBorrowChecker) CountByEquipmentAndStatusInTx(tx *gorm.DB, equipID uint, status string) (int64, error) {
	args := m.Called(tx, equipID, status)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockBorrowChecker) CountByEquipmentAndStatus(equipID uint, status string) (int64, error) {
	args := m.Called(equipID, status)
	return args.Get(0).(int64), args.Error(1)
}

// Ensure mock implements the interface
var _ service.BorrowStatusChecker = (*MockBorrowChecker)(nil)

func setupEquipService(t *testing.T) (*service.EquipmentService, *MockEquipDAO, *MockBorrowChecker, *miniredis.Miniredis) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	equipDAO := new(MockEquipDAO)
	borrowChecker := new(MockBorrowChecker)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	svc := service.NewEquipmentService(db, equipDAO, borrowChecker, rdb, zap.NewNop())
	return svc, equipDAO, borrowChecker, mr
}

// ──────────────────────── Create Tests ────────────────────────

func TestEquipService_Create_Success(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equipDAO.On("Insert", mock.AnythingOfType("*models.LabEquipment")).Return(nil)

	equip, err := svc.Create(context.Background(), &service.CreateEquipReq{
		Name:       "GPU Server",
		Model:      "RTX4090",
		Category:   "服务器",
		TotalStock: 10,
		Location:   "A101",
	})

	require.NoError(t, err)
	assert.Equal(t, "GPU Server", equip.Name)
	assert.Equal(t, uint(10), equip.TotalStock)
	assert.Equal(t, uint(10), equip.AvailableStock, "available should equal total on creation")
	assert.Equal(t, int8(1), equip.Status)
}

func TestEquipService_Create_ZeroStock(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equipDAO.On("Insert", mock.AnythingOfType("*models.LabEquipment")).Return(nil)

	equip, err := svc.Create(context.Background(), &service.CreateEquipReq{
		Name:       "Empty Device",
		TotalStock: 0,
	})

	require.NoError(t, err)
	assert.Equal(t, uint(0), equip.TotalStock)
	assert.Equal(t, uint(0), equip.AvailableStock)
}

// ──────────────────────── GetByID Tests ────────────────────────

func TestEquipService_GetByID_Success(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equip := &models.LabEquipment{
		ID: 1, Name: "GPU", Model: "A100", Category: "服务器",
		TotalStock: 10, AvailableStock: 8, Location: "A101", Status: 1,
	}
	equipDAO.On("FindByID", uint(1)).Return(equip, nil)

	dto, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "GPU", dto.Name)
	assert.Equal(t, uint(8), dto.AvailableStock)
}

func TestEquipService_GetByID_NotFound(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equipDAO.On("FindByID", uint(9999)).Return(nil, nil)

	dto, err := svc.GetByID(context.Background(), 9999)
	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "设备不存在")
}

func TestEquipService_GetByID_CacheHit(t *testing.T) {
	svc, equipDAO, _, mr := setupEquipService(t)

	// Pre-populate cache (时间使用 RFC3339 格式，与 time.Time JSON 序列化一致)
	mr.Set("equip:detail:1", `{"id":1,"name":"Cached GPU","model":"A100","category":"服务器","total_stock":10,"available_stock":8,"location":"A101","description":"","status":1,"created_at":"2026-06-13T12:00:00Z","updated_at":"2026-06-13T12:00:00Z"}`)

	dto, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Cached GPU", dto.Name)
	// DAO should NOT have been called
	equipDAO.AssertNotCalled(t, "FindByID")
}

// ──────────────────────── ListPage Tests ────────────────────────

func TestEquipService_ListPage_Success(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equips := []models.LabEquipment{
		{ID: 1, Name: "GPU", TotalStock: 5, AvailableStock: 5, Status: 1},
		{ID: 2, Name: "CPU", TotalStock: 3, AvailableStock: 3, Status: 1},
	}
	equipDAO.On("FindPage", 0, 12, "", "", (*int)(nil), false).Return(equips, int64(2), nil)

	result, err := svc.ListPage(context.Background(), &service.ListEquipReq{Page: 1, PageSize: 12})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.List, 2)
}

func TestEquipService_ListPage_RefreshFromDB(t *testing.T) {
	// When cache is empty, service should query DAO
	svc, equipDAO, _, _ := setupEquipService(t)

	equips := []models.LabEquipment{
		{ID: 1, Name: "GPU", TotalStock: 5, AvailableStock: 5, Status: 1},
	}
	equipDAO.On("FindPage", 0, 12, "", "", (*int)(nil), false).Return(equips, int64(1), nil)

	result, err := svc.ListPage(context.Background(), &service.ListEquipReq{Page: 1, PageSize: 12})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}

// ──────────────────────── Update Tests ────────────────────────

func TestEquipService_Update_NameOnly(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equip := &models.LabEquipment{
		ID: 1, Name: "Old Name", TotalStock: 10, AvailableStock: 8, Status: 1,
	}
	equipDAO.On("FindByID", uint(1)).Return(equip, nil)
	equipDAO.On("Update", equip).Return(nil)

	err := svc.Update(context.Background(), 1, &service.UpdateEquipReq{
		Name: "New Name",
	})
	require.NoError(t, err)
}

func TestEquipService_Update_TotalStockIncrease(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equip := &models.LabEquipment{
		ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 8, Status: 1,
	}
	newTotal := uint(15)
	equipDAO.On("FindByID", uint(1)).Return(equip, nil)
	equipDAO.On("Update", equip).Return(nil)

	err := svc.Update(context.Background(), 1, &service.UpdateEquipReq{
		TotalStock: &newTotal,
	})
	require.NoError(t, err)
	// available should increase: 8 + (15-10) = 13
	assert.Equal(t, uint(13), equip.AvailableStock)
}

func TestEquipService_Update_TotalStockDecrease(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equip := &models.LabEquipment{
		ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 8, Status: 1,
	}
	newTotal := uint(5)
	equipDAO.On("FindByID", uint(1)).Return(equip, nil)
	equipDAO.On("Update", equip).Return(nil)

	err := svc.Update(context.Background(), 1, &service.UpdateEquipReq{
		TotalStock: &newTotal,
	})
	require.NoError(t, err)
	// available should decrease: 8 + (5-10) = 3
	assert.Equal(t, uint(3), equip.AvailableStock)
}

func TestEquipService_Update_TotalStockCausesNegativeAvailable(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equip := &models.LabEquipment{
		ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 5, Status: 1,
	}
	newTotal := uint(3)
	equipDAO.On("FindByID", uint(1)).Return(equip, nil)

	err := svc.Update(context.Background(), 1, &service.UpdateEquipReq{
		TotalStock: &newTotal,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "库存不足")
}

func TestEquipService_Update_TotalStockExactZeroAvailable(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equip := &models.LabEquipment{
		ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 5, Status: 1,
	}
	newTotal := uint(5) // delta = -5, new_available = 5 + (-5) = 0
	equipDAO.On("FindByID", uint(1)).Return(equip, nil)
	equipDAO.On("Update", equip).Return(nil)

	err := svc.Update(context.Background(), 1, &service.UpdateEquipReq{
		TotalStock: &newTotal,
	})
	require.NoError(t, err)
	assert.Equal(t, uint(0), equip.AvailableStock)
}

func TestEquipService_Update_NotFound(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equipDAO.On("FindByID", uint(9999)).Return(nil, nil)

	err := svc.Update(context.Background(), 9999, &service.UpdateEquipReq{Name: "X"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "设备不存在")
}

// ──────────────────────── Disable Tests ────────────────────────

func TestEquipService_Disable_Success(t *testing.T) {
	svc, equipDAO, borrowChecker, _ := setupEquipService(t)

	equip := &models.LabEquipment{ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 10, Status: 1}
	equipDAO.On("GetForUpdate", mock.AnythingOfType("*gorm.DB"), uint(1)).Return(equip, nil)
	borrowChecker.On("CountByEquipmentAndStatusInTx", mock.AnythingOfType("*gorm.DB"), uint(1), "已借出").Return(int64(0), nil)
	equipDAO.On("UpdateWithTx", mock.AnythingOfType("*gorm.DB"), equip).Return(nil)

	err := svc.Disable(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int8(0), equip.Status)
}

func TestEquipService_Disable_HasActiveBorrows(t *testing.T) {
	svc, equipDAO, borrowChecker, _ := setupEquipService(t)

	equip := &models.LabEquipment{ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 8, Status: 1}
	equipDAO.On("GetForUpdate", mock.AnythingOfType("*gorm.DB"), uint(1)).Return(equip, nil)
	borrowChecker.On("CountByEquipmentAndStatusInTx", mock.AnythingOfType("*gorm.DB"), uint(1), "已借出").Return(int64(2), nil)

	err := svc.Disable(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法下架")
}

func TestEquipService_Disable_AlreadyDisabled(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equip := &models.LabEquipment{ID: 1, Name: "GPU", TotalStock: 10, AvailableStock: 10, Status: 0}
	equipDAO.On("GetForUpdate", mock.AnythingOfType("*gorm.DB"), uint(1)).Return(equip, nil)

	err := svc.Disable(context.Background(), 1)
	require.NoError(t, err, "disabling already disabled equipment should be idempotent")
}

func TestEquipService_Disable_NotFound(t *testing.T) {
	svc, equipDAO, _, _ := setupEquipService(t)

	equipDAO.On("GetForUpdate", mock.AnythingOfType("*gorm.DB"), uint(9999)).Return(nil, nil)

	err := svc.Disable(context.Background(), 9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "设备不存在")
}
