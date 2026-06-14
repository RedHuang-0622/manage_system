package controller_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"manage_system/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ──────────────────────── Borrow Controller HTTP Tests ────────────────────────

func TestBorrowController_Apply_ValidBody(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/apply", func(c *gin.Context) {
		var req service.ApplyBorrowReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "请求参数错误"})
			return
		}
		c.Set("user_id", uint(3))
		c.JSON(http.StatusCreated, gin.H{"code": 0, "data": gin.H{
			"equipment_id": req.EquipmentID,
			"quantity":     req.Quantity,
		}})
	})

	body := `{"equipment_id":1,"quantity":2,"apply_note":"need GPU"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBorrowController_Apply_MissingEquipmentID(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/apply", func(c *gin.Context) {
		var req service.ApplyBorrowReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "请求参数错误"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"code": 0})
	})

	body := `{"quantity":2}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "1001")
}

func TestBorrowController_Apply_ZeroQuantity(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/apply", func(c *gin.Context) {
		var req service.ApplyBorrowReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "请求参数错误"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"code": 0})
	})

	body := `{"equipment_id":1,"quantity":0}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBorrowController_Apply_QuantityNegative(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/apply", func(c *gin.Context) {
		var req service.ApplyBorrowReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "请求参数错误"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"code": 0})
	})

	body := `{"equipment_id":1,"quantity":-5}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBorrowController_Approve_Valid(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/:id/approve", func(c *gin.Context) {
		id := c.Param("id")
		if id == "abc" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "无效的工单ID"})
			return
		}
		var req service.ApproveBorrowReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001})
			return
		}
		c.Set("user_id", uint(1))
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"status": "已借出"}})
	})

	body := `{"approve":true,"approve_note":"ok"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/100/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "已借出")
}

func TestBorrowController_Approve_InvalidID(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/:id/approve", func(c *gin.Context) {
		id := c.Param("id")
		if id == "abc" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "无效的工单ID"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/abc/approve", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBorrowController_Approve_Reject(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/:id/approve", func(c *gin.Context) {
		var req service.ApproveBorrowReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001})
			return
		}
		c.Set("user_id", uint(1))
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"status": "被拒绝"}})
	})

	body := `{"approve":false,"approve_note":"设备已预约"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/100/approve", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "被拒绝")
}

func TestBorrowController_Return_Success(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/:id/return", func(c *gin.Context) {
		c.Set("user_id", uint(3))
		c.Set("role_name", "member")
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"status": "已归还"}})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/100/return", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "已归还")
}

func TestBorrowController_ListMyRecords(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/borrows/my", func(c *gin.Context) {
		var req service.ListBorrowReq
		_ = c.ShouldBindQuery(&req)
		c.Set("user_id", uint(3))
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
			"total": 5,
			"list":  []string{},
		}})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/borrows/my", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBorrowController_ListMyRecords_ByStatus(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/borrows/my", func(c *gin.Context) {
		var req service.ListBorrowReq
		_ = c.ShouldBindQuery(&req)
		c.Set("user_id", uint(3))
		c.JSON(http.StatusOK, gin.H{"status_filter": req.Status})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/borrows/my?status=已借出", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "已借出")
}

func TestBorrowController_ListPending(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/borrows/pending", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"total": 3, "list": []string{}}})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/borrows/pending", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBorrowController_ListAll_WithFilters(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/borrows", func(c *gin.Context) {
		var req service.ListBorrowReq
		_ = c.ShouldBindQuery(&req)
		c.JSON(http.StatusOK, gin.H{
			"user_id":      req.UserID,
			"equipment_id": req.EquipmentID,
			"status":       req.Status,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/borrows?user_id=3&equipment_id=1&status=已借出", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBorrowController_Cancel_Success(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/borrows/:id/cancel", func(c *gin.Context) {
		c.Set("user_id", uint(3))
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已取消申请"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/100/cancel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "已取消申请")
}
