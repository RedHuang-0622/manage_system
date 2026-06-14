package pkg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"manage_system/pkg/errcode"
	"manage_system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

func TestResponse_Success_WithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Success(c, gin.H{"name": "test"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t,
		`{"code":0,"msg":"success","data":{"name":"test"}}`,
		w.Body.String(),
	)
}

func TestResponse_Success_WithNil(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Success(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t,
		`{"code":0,"msg":"success","data":null}`,
		w.Body.String(),
	)
}

func TestResponse_Created(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Created(c, gin.H{"id": 1})

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t,
		`{"code":0,"msg":"创建成功","data":{"id":1}}`,
		w.Body.String(),
	)
}

func TestResponse_Error_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Error(c, http.StatusBadRequest, errcode.ErrInvalidParam, "请求参数错误")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t,
		`{"code":1001,"msg":"请求参数错误","data":null}`,
		w.Body.String(),
	)
}

func TestResponse_Error_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Error(c, http.StatusUnauthorized, errcode.ErrAuthFailed, "用户名或密码错误")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t,
		`{"code":2001,"msg":"用户名或密码错误","data":null}`,
		w.Body.String(),
	)
}

func TestResponse_Error_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Error(c, http.StatusForbidden, errcode.ErrPermissionDenied, "权限不足")

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t,
		`{"code":2005,"msg":"权限不足","data":null}`,
		w.Body.String(),
	)
}

func TestResponse_Error_InternalServer(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.Error(c, http.StatusInternalServerError, errcode.ErrInternalServer, "系统内部错误")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t,
		`{"code":5000,"msg":"系统内部错误","data":null}`,
		w.Body.String(),
	)
}

func TestPageRequest_Normalize_Defaults(t *testing.T) {
	p := &response.PageRequest{}
	offset, limit := p.Normalize()

	assert.Equal(t, 0, offset)
	assert.Equal(t, 10, limit)
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 10, p.PageSize)
}

func TestPageRequest_Normalize_ValidInput(t *testing.T) {
	p := &response.PageRequest{Page: 2, PageSize: 20}
	offset, limit := p.Normalize()

	assert.Equal(t, 20, offset)
	assert.Equal(t, 20, limit)
}

func TestPageRequest_Normalize_ClampMax(t *testing.T) {
	p := &response.PageRequest{Page: 1, PageSize: 200}
	_, limit := p.Normalize()

	assert.Equal(t, 100, limit)
}

func TestPageRequest_Normalize_ZeroPage(t *testing.T) {
	p := &response.PageRequest{Page: 0, PageSize: 5}
	offset, limit := p.Normalize()

	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 0, offset)
	assert.Equal(t, 5, limit)
}

func TestPageResponse_Normal(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.PageResponse(c, 50, 2, 10, []string{"a", "b"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t,
		`{"code":0,"msg":"success","data":{"total":50,"page":2,"page_size":10,"list":["a","b"]}}`,
		w.Body.String(),
	)
}

func TestPageResponse_EmptyList(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	response.PageResponse(c, 0, 1, 10, []string{})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t,
		`{"code":0,"msg":"success","data":{"total":0,"page":1,"page_size":10,"list":[]}}`,
		w.Body.String(),
	)
}
