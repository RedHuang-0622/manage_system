package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"manage_system/pkg/errcode"
	resp "manage_system/pkg/response"
	"manage_system/router/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() { gin.SetMode(gin.TestMode) }

func setupRecoveryTest(t *testing.T) (*gin.Engine, *bytes.Buffer) {
	t.Helper()

	buf := &bytes.Buffer{}
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:     "ts",
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeTime:  zapcore.ISO8601TimeEncoder,
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(middleware.Recovery(logger))

	return router, buf
}

func TestRecovery_NormalRequestNoPanic(t *testing.T) {
	router, _ := setupRecoveryTest(t)

	router.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, resp.Response{Code: 0, Msg: "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ok", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecovery_PanicCaptured(t *testing.T) {
	router, buf := setupRecoveryTest(t)

	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "5000")
	assert.Contains(t, w.Body.String(), errcode.GetMsg(errcode.ErrInternalServer))
	assert.Contains(t, buf.String(), "panic_recovered")
	assert.Contains(t, buf.String(), "test panic")
}

func TestRecovery_PanicLogContainsRequestInfo(t *testing.T) {
	router, buf := setupRecoveryTest(t)

	router.POST("/api/v1/equipments", func(c *gin.Context) {
		panic("equipment panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/equipments", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	router.ServeHTTP(w, req)

	assert.Contains(t, buf.String(), "panic_recovered")
	assert.Contains(t, buf.String(), "path")
	assert.Contains(t, buf.String(), "method")
	assert.Contains(t, buf.String(), "stack")
}

func TestRecovery_PanicTerminatesRequest(t *testing.T) {
	router, _ := setupRecoveryTest(t)

	executed := false
	router.GET("/panic-chain", func(c *gin.Context) {
		panic("stop here")
		executed = true // unreachable
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic-chain", nil)
	router.ServeHTTP(w, req)

	assert.False(t, executed, "code after panic should not execute")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecovery_NestedPanic(t *testing.T) {
	router, buf := setupRecoveryTest(t)

	router.GET("/nested", func(c *gin.Context) {
		func() {
			panic("nested panic")
		}()
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nested", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, buf.String(), "nested panic")
}
