package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"manage_system/router/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func requireNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

const aclModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
`

func setupCasbinTest(t *testing.T, policies [][]string) (*gin.Engine, *bytes.Buffer) {
	t.Helper()

	m, err := model.NewModelFromString(aclModel)
	requireNoErr(t, err)
	enforcer, err := casbin.NewEnforcer(m)
	requireNoErr(t, err)

	for _, p := range policies {
		if len(p) == 3 {
			enforcer.AddPolicy(p[0], p[1], p[2])
		} else if len(p) == 2 {
			enforcer.AddGroupingPolicy(p[0], p[1])
		}
	}

	buf := &bytes.Buffer{}
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:     "ts",
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeTime:  zapcore.ISO8601TimeEncoder,
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.WarnLevel)
	logger := zap.New(core)

	router := gin.New()

	// Fake auth middleware: reads role from header, sets it in context BEFORE Casbin runs
	router.Use(func(c *gin.Context) {
		if role := c.GetHeader("X-Test-Role"); role != "" {
			c.Set("role_name", role)
		}
		if uidStr := c.GetHeader("X-Test-UserID"); uidStr != "" {
			uid, _ := strconv.ParseUint(uidStr, 10, 64)
			c.Set("user_id", uint(uid))
		}
		// Always set a default username to avoid nil pointer in Casbin middleware logging
		username := c.GetHeader("X-Test-Username")
		if username == "" {
			username = "testuser"
		}
		c.Set("username", username)
		c.Next()
	})

	router.Use(middleware.Casbin(enforcer, logger))

	return router, buf
}

func TestCasbinMiddleware_Allowed(t *testing.T) {
	policies := [][]string{
		{"member", "/api/v1/equipments", "GET"},
	}
	router, _ := setupCasbinTest(t, policies)

	router.GET("/api/v1/equipments", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("X-Test-Role", "member")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCasbinMiddleware_Denied(t *testing.T) {
	policies := [][]string{
		{"member", "/api/v1/equipments", "GET"},
	}
	router, buf := setupCasbinTest(t, policies)

	router.POST("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", nil)
	req.Header.Set("X-Test-Role", "member")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "2005") // ErrPermissionDenied
	assert.Contains(t, buf.String(), "permission_denied")
}

func TestCasbinMiddleware_LabAdmin(t *testing.T) {
	policies := [][]string{
		{"lab_admin", "/api/v1/users*", ".*"},
		{"lab_admin", "/api/v1/equipments*", ".*"},
	}
	router, _ := setupCasbinTest(t, policies)

	// lab_admin can manage users
	router.POST("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", nil)
	req.Header.Set("X-Test-Role", "lab_admin")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// lab_admin can manage equipments
	router.PUT("/api/v1/equipments/1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("PUT", "/api/v1/equipments/1", nil)
	req2.Header.Set("X-Test-Role", "lab_admin")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestCasbinMiddleware_SuperAdminFullAccess(t *testing.T) {
	policies := [][]string{
		{"super_admin", "/api/v1/*", ".*"},
	}
	router, _ := setupCasbinTest(t, policies)

	router.POST("/api/v1/borrows/1/approve", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/1/approve", nil)
	req.Header.Set("X-Test-Role", "super_admin")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCasbinMiddleware_NoRoleName(t *testing.T) {
	router, _ := setupCasbinTest(t, nil)

	router.GET("/api/v1/equipments", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	// No X-Test-Role header set
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "2005")
}

func TestCasbinMiddleware_MemberApproveDenied(t *testing.T) {
	policies := [][]string{
		{"member", "/api/v1/equipments*", "GET"},
		{"member", "/api/v1/borrows/apply", "POST"},
	}
	router, buf := setupCasbinTest(t, policies)

	router.POST("/api/v1/borrows/1/approve", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/1/approve", nil)
	req.Header.Set("X-Test-Role", "member")
	req.Header.Set("X-Test-UserID", "3")
	req.Header.Set("X-Test-Username", "member1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, buf.String(), "permission_denied")
	assert.Contains(t, buf.String(), "member")
	assert.Contains(t, buf.String(), "/api/v1/borrows/1/approve")
}

func TestCasbinMiddleware_GlobMatching(t *testing.T) {
	// keyMatch supports glob patterns with *, not regex \d+
	policies := [][]string{
		{"member", "/api/v1/borrows/*/return", "POST"},
	}
	router, _ := setupCasbinTest(t, policies)

	router.POST("/api/v1/borrows/123/return", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/123/return", nil)
	req.Header.Set("X-Test-Role", "member")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCasbinMiddleware_MemberCanApply(t *testing.T) {
	policies := [][]string{
		{"member", "/api/v1/borrows/apply", "POST"},
	}
	router, _ := setupCasbinTest(t, policies)

	router.POST("/api/v1/borrows/apply", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/borrows/apply", nil)
	req.Header.Set("X-Test-Role", "member")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}
