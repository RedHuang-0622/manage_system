package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtpkg "manage_system/pkg/jwt"
	"manage_system/pkg/config"
	"manage_system/router/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthMiddleware(t *testing.T) (*gin.Engine, *jwtpkg.Service, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := config.JWTConfig{
		Secret: "test-secret-key-at-least-32-chars-long!!",
		Expire: 7200,
		Issuer: "test-issuer",
	}
	jwtService := jwtpkg.NewService(cfg, rdb)

	router := gin.New()
	router.Use(middleware.Auth(jwtService))

	router.GET("/api/v1/equipments", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":   c.GetUint("user_id"),
			"username":  c.GetString("username"),
			"role_name": c.GetString("role_name"),
		})
	})

	return router, jwtService, mr
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router, jwtService, _ := setupAuthMiddleware(t)

	token, _, err := jwtService.GenerateToken(1, "testuser", 3, "member")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":1`)
	assert.Contains(t, w.Body.String(), `"username":"testuser"`)
	assert.Contains(t, w.Body.String(), `"role_name":"member"`)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router, _, _ := setupAuthMiddleware(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "2003") // ErrTokenMissing
}

func TestAuthMiddleware_WrongFormat(t *testing.T) {
	router, _, _ := setupAuthMiddleware(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "2003")
}

func TestAuthMiddleware_BlacklistedToken(t *testing.T) {
	router, jwtService, _ := setupAuthMiddleware(t)

	token, _, err := jwtService.GenerateToken(1, "testuser", 3, "member")
	require.NoError(t, err)

	// Add to blacklist
	err = jwtService.AddToBlacklist(token, time.Now().Add(time.Hour))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token已失效")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := config.JWTConfig{
		Secret: "test-secret-key-at-least-32-chars-long!!",
		Expire: -1, // already expired
		Issuer: "test-issuer",
	}
	jwtService := jwtpkg.NewService(cfg, rdb)

	router := gin.New()
	router.Use(middleware.Auth(jwtService))
	router.GET("/api/v1/equipments", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token, _, err := jwtService.GenerateToken(1, "testuser", 3, "member")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "2004") // ErrTokenInvalid
}

func TestAuthMiddleware_TamperedToken(t *testing.T) {
	router, _, _ := setupAuthMiddleware(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/equipments", nil)
	req.Header.Set("Authorization", "Bearer tampered.token.string")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "2004") // ErrTokenInvalid
}

func TestAuthMiddleware_WhiteListSkip(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := config.JWTConfig{
		Secret: "test-secret-key-at-least-32-chars-long!!",
		Expire: 7200,
		Issuer: "test-issuer",
	}
	jwtService := jwtpkg.NewService(cfg, rdb)

	router := gin.New()
	router.Use(middleware.Auth(jwtService))

	// Health endpoint
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Login endpoint
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"login": "ok"})
	})

	// Health check should work without token
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Login should work without token
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
