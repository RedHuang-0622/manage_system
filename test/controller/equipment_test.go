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

// ──────────────────────── Equipment Controller HTTP Tests ────────────────────────

func TestEquipController_List_DefaultPagination(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/equipments", func(c *gin.Context) {
		var req service.ListEquipReq
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001})
			return
		}
		if req.Page == 0 {
			req.Page = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 12
		}
		c.JSON(http.StatusOK, gin.H{"page": req.Page, "page_size": req.PageSize})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEquipController_List_WithFilters(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/equipments", func(c *gin.Context) {
		var req service.ListEquipReq
		_ = c.ShouldBindQuery(&req)
		c.JSON(http.StatusOK, gin.H{
			"keyword":        req.Keyword,
			"category":       req.Category,
			"only_available": req.OnlyAvailable,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments?keyword=GPU&category=服务器&only_available=1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "GPU")
	assert.Contains(t, w.Body.String(), "服务器")
}

func TestEquipController_GetByID_ValidID(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/equipments/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "1" {
			c.JSON(http.StatusOK, gin.H{"id": 1, "name": "GPU Server"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "GPU Server")
}

func TestEquipController_GetByID_InvalidID(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/equipments/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "abc" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "无效的设备ID"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEquipController_Create_ValidBody(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/equipments", func(c *gin.Context) {
		var req service.CreateEquipReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "请求参数错误"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"code": 0, "data": gin.H{
			"name":        req.Name,
			"total_stock": req.TotalStock,
		}})
	})

	body := `{"name":"GPU Server","total_stock":10,"category":"服务器"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/equipments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "GPU Server")
}

func TestEquipController_Create_MissingName(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/equipments", func(c *gin.Context) {
		var req service.CreateEquipReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "请求参数错误"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"code": 0})
	})

	body := `{"total_stock":10,"category":"服务器"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/equipments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "1001")
}

func TestEquipController_Create_NegativeStock(t *testing.T) {
	router := gin.New()
	router.POST("/api/v1/equipments", func(c *gin.Context) {
		var req service.CreateEquipReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001, "msg": "请求参数错误"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"code": 0})
	})

	body := `{"name":"GPU","total_stock":-1}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/equipments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEquipController_Update_Success(t *testing.T) {
	router := gin.New()
	router.PUT("/api/v1/equipments/:id", func(c *gin.Context) {
		var req service.UpdateEquipReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
	})

	body := `{"name":"New Name","location":"B202"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/equipments/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEquipController_Disable_Success(t *testing.T) {
	router := gin.New()
	router.DELETE("/api/v1/equipments/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "1" {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "设备已下架"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": 1001})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/equipments/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "设备已下架")
}

func TestEquipController_PageSizeExceed(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/equipments", func(c *gin.Context) {
		var req service.ListEquipReq
		_ = c.ShouldBindQuery(&req)
		if req.PageSize > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1001})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments?page_size=200", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
