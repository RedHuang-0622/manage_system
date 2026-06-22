//go:build integration
// +build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"manage_system/controller"
	"manage_system/dao"
	"manage_system/models"
	"manage_system/pkg/config"
	jwtpkg "manage_system/pkg/jwt"
	"manage_system/pkg/response"
	"manage_system/router"
	"manage_system/service"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() { gin.SetMode(gin.TestMode) }

// ──────────────────────── Test Harness ────────────────────────

type TestHarness struct {
	DB          *gorm.DB
	Redis       *redis.Client
	Router      *gin.Engine
	AdminToken  string
	MemberToken string
	AdminID     uint
	MemberID    uint
	EquipID     uint
	BorrowID    uint
}

func setupIntegrationTest(t *testing.T) *TestHarness {
	t.Helper()

	// Setup SQLite in-memory
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Setup Casbin model
	casbinModel, err := casbinmodel.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
`)
	require.NoError(t, err)
	casbinEnforcer, err := casbin.NewEnforcer(casbinModel)
	require.NoError(t, err)

	// Setup JWT
	jwtService := jwtpkg.NewService(config.JWTConfig{
		Secret: "integration-test-secret-key-32chars!!",
		Expire: 7200,
		Issuer: "integration-test",
	}, nil) // nil redis for integration tests

	// AutoMigrate
	err = db.AutoMigrate(
		&models.SysUser{},
		&models.SysRole{},
		&models.LabEquipment{},
		&models.BorrowRecord{},
	)
	require.NoError(t, err)

	// Seed roles
	db.Create(&models.SysRole{ID: 1, RoleName: "super_admin", Description: "超级管理员", IsSystem: 1})
	db.Create(&models.SysRole{ID: 2, RoleName: "lab_admin", Description: "实验室负责人", IsSystem: 1})
	db.Create(&models.SysRole{ID: 3, RoleName: "member", Description: "普通成员", IsSystem: 1})

	// Seed admin user
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 12)
	db.Create(&models.SysUser{
		Username: "admin", PasswordHash: string(adminHash),
		RealName: "管理员", RoleID: 1, Status: 1,
	})

	// Seed Casbin policies
	policies := [][]string{
		{"super_admin", "/api/v1/*", ".*"},
		{"lab_admin", "/api/v1/users*", ".*"},
		{"lab_admin", "/api/v1/equipments*", ".*"},
		{"lab_admin", "/api/v1/borrows*", ".*"},
		{"lab_admin", "/api/v1/roles*", "GET"},
		{"lab_admin", "/api/v1/auth/logout", "POST"},
		{"lab_admin", "/api/v1/auth/refresh", "POST"},
		{"member", "/api/v1/equipments*", "GET"},
		{"member", "/api/v1/borrows/apply", "POST"},
		{"member", "/api/v1/borrows/my", "GET"},
		{"member", "/api/v1/borrows/\\d+/return", "POST"},
		{"member", "/api/v1/borrows/\\d+/cancel", "POST"},
		{"member", "/api/v1/users/\\d+/password", "PUT"},
		{"member", "/api/v1/roles", "GET"},
		{"member", "/api/v1/auth/logout", "POST"},
		{"member", "/api/v1/auth/refresh", "POST"},
	}
	for _, p := range policies {
		casbinEnforcer.AddPolicy(p[0], p[1], p[2])
	}
	casbinEnforcer.AddGroupingPolicy("member", "member")
	casbinEnforcer.AddGroupingPolicy("lab_admin", "lab_admin")
	casbinEnforcer.AddGroupingPolicy("super_admin", "super_admin")
	casbinEnforcer.AddGroupingPolicy("super_admin", "lab_admin")

	// Build dependencies
	userDAO := dao.NewUserDAO(db)
	roleDAO := dao.NewRoleDAO(db)
	equipDAO := dao.NewEquipmentDAO(db)
	borrowDAO := dao.NewBorrowDAO(db)

	iamService := service.NewIAMService(db, userDAO, roleDAO, jwtService, nil)
	equipService := service.NewEquipmentService(db, equipDAO, borrowDAO, nil, nil)
	borrowService := service.NewBorrowService(db, borrowDAO, equipDAO, equipService, nil)

	r := router.SetupRouter(router.Dependencies{
		AuthController:      controller.NewAuthController(iamService),
		EquipmentController: controller.NewEquipmentController(equipService),
		BorrowController:    controller.NewBorrowController(borrowService),
		JWTService:          jwtService,
		Enforcer:            casbinEnforcer,
		Logger:              nil,
	})

	// Login as admin
	adminToken := loginHelper(t, r, "admin", "admin123")

	// Create member user and login
	createUserHelper(t, r, adminToken, "member1", "password123", "成员1", 3)
	memberToken := loginHelper(t, r, "member1", "password123")

	// Create equipment as admin
	equipID := createEquipHelper(t, r, adminToken, "GPU Server", "RTX4090", "服务器", 10, "A101")

	return &TestHarness{
		DB:          db,
		Router:      r,
		AdminToken:  adminToken,
		MemberToken: memberToken,
		AdminID:     1,
		MemberID:    2,
		EquipID:     equipID,
	}
}

// ──────────────────────── Helper Functions ────────────────────────

func loginHelper(t *testing.T, r *gin.Engine, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	return data["token"].(string)
}

func createUserHelper(t *testing.T, r *gin.Engine, adminToken, username, password, realName string, roleID uint) {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"username":  username,
		"password":  password,
		"real_name": realName,
		"role_id":   roleID,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func createEquipHelper(t *testing.T, r *gin.Engine, adminToken, name, model, category string, totalStock uint, location string) uint {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"name":        name,
		"model":       model,
		"category":    category,
		"total_stock": totalStock,
		"location":    location,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/equipments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	return uint(data["id"].(float64))
}

// ──────────────────────── Scenario 1: Full Borrow Flow ────────────────────────

func TestIntegration_FullBorrowFlow(t *testing.T) {
	h := setupIntegrationTest(t)

	// Step 1: Create another member
	createUserHelper(t, h.Router, h.AdminToken, "user2", "pass123456", "用户2", 3)

	// Step 2: Member login
	memberToken := loginHelper(t, h.Router, "user2", "pass123456")

	// Step 3: View equipment list as member
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	t.Log("✓ Equipment list accessible for members")

	// Step 4: Apply for borrow
	applyBody, _ := json.Marshal(map[string]interface{}{
		"equipment_id": h.EquipID,
		"quantity":     2,
		"apply_note":   "Need for experiment",
	})

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBuffer(applyBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	borrowID := uint(data["id"].(float64))
	assert.Equal(t, "申请中", data["status"])
	t.Logf("✓ Borrow applied: ID=%d, status=%s", borrowID, data["status"])

	// Step 5: Verify stock unchanged (apply doesn't deduct)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/equipments/"+formatUint(h.EquipID), nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	equipData := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(10), equipData["available_stock"])
	t.Logf("✓ Stock unchanged after apply: available=%v", equipData["available_stock"])

	// Step 6: Admin approve
	approveBody, _ := json.Marshal(map[string]interface{}{
		"approve":      true,
		"approve_note": "Approved",
	})

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowID)+"/approve", bytes.NewBuffer(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	data = resp.Data.(map[string]interface{})
	assert.Equal(t, "已借出", data["status"])
	t.Logf("✓ Borrow approved: status=%s", data["status"])

	// Step 7: Verify stock deducted
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/equipments/"+formatUint(h.EquipID), nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	equipData = resp.Data.(map[string]interface{})
	assert.Equal(t, float64(8), equipData["available_stock"]) // 10 - 2
	t.Logf("✓ Stock deducted: available=%v", equipData["available_stock"])

	// Step 8: Return
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowID)+"/return", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	data = resp.Data.(map[string]interface{})
	assert.Equal(t, "已归还", data["status"])
	t.Logf("✓ Returned: status=%s", data["status"])

	// Step 9: Verify stock restored
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/equipments/"+formatUint(h.EquipID), nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	equipData = resp.Data.(map[string]interface{})
	assert.Equal(t, float64(10), equipData["available_stock"])
	t.Logf("✓ Stock restored: available=%v", equipData["available_stock"])

	// Step 10: Verify my borrow records
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/borrows/my", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	pageData := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), pageData["total"])
	t.Log("✓ My borrow records: total=1")
}

// ──────────────────────── Scenario 2: Approve Reject Flow ────────────────────────

func TestIntegration_ApproveRejectFlow(t *testing.T) {
	h := setupIntegrationTest(t)

	// Apply
	applyBody, _ := json.Marshal(map[string]interface{}{
		"equipment_id": h.EquipID,
		"quantity":     1,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBuffer(applyBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	borrowID := uint(data["id"].(float64))

	// Reject
	rejectBody, _ := json.Marshal(map[string]interface{}{
		"approve":      false,
		"approve_note": "设备已预约",
	})

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowID)+"/approve", bytes.NewBuffer(rejectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	data = resp.Data.(map[string]interface{})
	assert.Equal(t, "被拒绝", data["status"])
	t.Logf("✓ Rejected: status=%s", data["status"])

	// Try to approve again (should fail — terminal state)
	approveBody, _ := json.Marshal(map[string]interface{}{"approve": true})

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowID)+"/approve", bytes.NewBuffer(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	t.Logf("✓ Cannot approve rejected borrow: HTTP %d", w.Code)

	// Verify stock unchanged (reject doesn't deduct)
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/equipments/"+formatUint(h.EquipID), nil)
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	equipData := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(10), equipData["available_stock"])
	t.Log("✓ Stock unchanged after reject")
}

// ──────────────────────── Scenario 3: Permission Boundary ────────────────────────

func TestIntegration_PermissionBoundary(t *testing.T) {
	h := setupIntegrationTest(t)

	// Member tries to create user — should be 403
	createUserBody, _ := json.Marshal(map[string]interface{}{
		"username":  "hacker",
		"password":  "password123",
		"real_name": "Hacker",
		"role_id":   3,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(createUserBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	t.Log("✓ Member cannot create users (403)")

	// Member tries to approve — should be 403
	approveBody, _ := json.Marshal(map[string]interface{}{"approve": true})

	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/1/approve", bytes.NewBuffer(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	t.Log("✓ Member cannot approve borrows (403)")

	// Member can view equipment
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	t.Log("✓ Member can view equipment (200)")
}

// ──────────────────────── Scenario 4: JWT Blacklist ────────────────────────

func TestIntegration_JWTBlacklist(t *testing.T) {
	h := setupIntegrationTest(t)

	// Logout
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	t.Log("✓ Logout successful")

	// Try to use the old token
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	t.Logf("✓ Old token rejected after logout: HTTP %d", w.Code)

	// Refresh token
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	newToken := data["token"].(string)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, h.AdminToken, newToken)
	t.Log("✓ Token refresh successful, new token issued")
}

// ──────────────────────── Scenario 5: Duplicate Borrow Prevention ────────────────────────

func TestIntegration_DuplicateBorrowPrevention(t *testing.T) {
	h := setupIntegrationTest(t)

	// First apply
	applyBody, _ := json.Marshal(map[string]interface{}{
		"equipment_id": h.EquipID,
		"quantity":     1,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBuffer(applyBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	t.Log("✓ First borrow application created")

	// Second apply (same user + equipment) — should fail
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBuffer(applyBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
	t.Log("✓ Duplicate borrow prevented (409)")
}

// ──────────────────────── Scenario 6: Disable Equipment Protection ────────────────────────

func TestIntegration_DisableEquipmentProtection(t *testing.T) {
	h := setupIntegrationTest(t)

	// Apply and approve a borrow
	applyBody, _ := json.Marshal(map[string]interface{}{
		"equipment_id": h.EquipID,
		"quantity":     1,
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBuffer(applyBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	borrowID := uint(data["id"].(float64))

	// Approve
	approveBody, _ := json.Marshal(map[string]interface{}{"approve": true})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowID)+"/approve", bytes.NewBuffer(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Try to disable equipment with active borrow
	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/v1/equipments/"+formatUint(h.EquipID), nil)
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
	t.Logf("✓ Cannot disable equipment with active borrow: HTTP %d", w.Code)

	// Return and retry
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowID)+"/return", nil)
	req.Header.Set("Authorization", "Bearer "+h.MemberToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/v1/equipments/"+formatUint(h.EquipID), nil)
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	t.Log("✓ Can disable equipment after all borrows returned")
}

// ──────────────────────── Scenario 7: Self Disable Prevention ────────────────────────

func TestIntegration_SelfDisablePrevention(t *testing.T) {
	h := setupIntegrationTest(t)

	// Admin tries to disable themselves
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users/1/disable", nil)
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	t.Log("✓ Admin cannot disable themselves")
}

// ──────────────────────── Scenario 8: Concurrent Approve ────────────────────────

func TestIntegration_ConcurrentApprove(t *testing.T) {
	h := setupIntegrationTest(t)

	// Create 5 borrow applications on the same equipment
	borrowIDs := make([]uint, 5)
	for i := 0; i < 5; i++ {
		createUserHelper(t, h.Router, h.AdminToken, "user"+formatUint(uint(i)), "pass123456", "User", 3)
		token := loginHelper(t, h.Router, "user"+formatUint(uint(i)), "pass123456")

		applyBody, _ := json.Marshal(map[string]interface{}{
			"equipment_id": h.EquipID,
			"quantity":     1,
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBuffer(applyBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		h.Router.ServeHTTP(w, req)

		var resp response.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp.Data.(map[string]interface{})
		borrowIDs[i] = uint(data["id"].(float64))
	}
	t.Logf("Created 5 borrow applications, IDs: %v", borrowIDs)

	// Concurrent approve
	var wg sync.WaitGroup
	results := make(chan int, 5) // 0=fail, 1=success

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			approveBody, _ := json.Marshal(map[string]interface{}{"approve": true})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowIDs[idx])+"/approve", bytes.NewBuffer(approveBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+h.AdminToken)
			h.Router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				results <- 1
			} else {
				results <- 0
			}
		}(i)
	}
	wg.Wait()
	close(results)

	successCount := 0
	for r := range results {
		successCount += r
	}

	// Available stock is 10, each borrow uses 1, so max 10 successes
	assert.LessOrEqual(t, successCount, 10)
	t.Logf("Concurrent approves: %d/%d succeeded", successCount, 5)

	// Verify final stock consistency
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments/"+formatUint(h.EquipID), nil)
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	equipData := resp.Data.(map[string]interface{})
	availStock := equipData["available_stock"].(float64)
	assert.Equal(t, float64(10)-float64(successCount), availStock,
		"available stock should equal initial minus successful approvals")
	t.Logf("Final available stock: %v (consistent)", availStock)
}

// ──────────────────────── Scenario 9: Stock Linkage Update ────────────────────────

func TestIntegration_StockLinkageOnUpdate(t *testing.T) {
	h := setupIntegrationTest(t)

	// Create equipment with total=10
	body, _ := json.Marshal(map[string]interface{}{
		"name":        "Linkage Test Equipment",
		"total_stock": 10,
		"category":    "测试",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/equipments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	equipID := uint(data["id"].(float64))

	// Apply + approve borrow (qty=3)
	createUserHelper(t, h.Router, h.AdminToken, "linkuser", "pass123456", "LinkUser", 3)
	memberToken := loginHelper(t, h.Router, "linkuser", "pass123456")

	applyBody, _ := json.Marshal(map[string]interface{}{"equipment_id": equipID, "quantity": 3})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBuffer(applyBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+memberToken)
	h.Router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	data = resp.Data.(map[string]interface{})
	borrowID := uint(data["id"].(float64))

	approveBody, _ := json.Marshal(map[string]interface{}{"approve": true})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/borrows/"+formatUint(borrowID)+"/approve", bytes.NewBuffer(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Now available = 7, total = 10

	// Update total from 10 to 8
	updateBody, _ := json.Marshal(map[string]interface{}{"total_stock": 8})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/v1/equipments/"+formatUint(equipID), bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	t.Log("✓ Updated total from 10 to 8")

	// Verify: available should be 7 + (8-10) = 5
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/equipments/"+formatUint(equipID), nil)
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &resp)
	equipData := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(8), equipData["total_stock"])
	assert.Equal(t, float64(5), equipData["available_stock"])
	t.Logf("✓ Stock linkage correct: total=%v, available=%v", equipData["total_stock"], equipData["available_stock"])

	// Try to set total below available (should fail)
	updateBody2, _ := json.Marshal(map[string]interface{}{"total_stock": 3})
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/v1/equipments/"+formatUint(equipID), bytes.NewBuffer(updateBody2))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.AdminToken)
	h.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	t.Log("✓ Cannot set total below available (stock protection)")
}

// ──────────────────────── Utility ────────────────────────

func formatUint(n uint) string {
	return formatUint64(uint64(n))
}

func formatUint64(n uint64) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
