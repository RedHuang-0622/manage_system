package pkg_test

import (
	"testing"

	"manage_system/pkg/errcode"

	"github.com/stretchr/testify/assert"
)

func TestGetMsg_Success(t *testing.T) {
	assert.Equal(t, "success", errcode.GetMsg(errcode.Success))
}

func TestGetMsg_InvalidParam(t *testing.T) {
	assert.Equal(t, "请求参数错误", errcode.GetMsg(errcode.ErrInvalidParam))
}

func TestGetMsg_AuthFailed(t *testing.T) {
	assert.Equal(t, "用户名或密码错误", errcode.GetMsg(errcode.ErrAuthFailed))
}

func TestGetMsg_AccountDisabled(t *testing.T) {
	assert.Equal(t, "账号已被禁用", errcode.GetMsg(errcode.ErrAccountDisabled))
}

func TestGetMsg_TokenMissing(t *testing.T) {
	assert.Equal(t, "未提供认证Token", errcode.GetMsg(errcode.ErrTokenMissing))
}

func TestGetMsg_TokenInvalid(t *testing.T) {
	assert.Equal(t, "Token无效或已过期", errcode.GetMsg(errcode.ErrTokenInvalid))
}

func TestGetMsg_PermissionDenied(t *testing.T) {
	assert.Equal(t, "权限不足", errcode.GetMsg(errcode.ErrPermissionDenied))
}

func TestGetMsg_UserExists(t *testing.T) {
	assert.Equal(t, "用户名已存在", errcode.GetMsg(errcode.ErrUserExists))
}

func TestGetMsg_UserNotFound(t *testing.T) {
	assert.Equal(t, "用户不存在", errcode.GetMsg(errcode.ErrUserNotFound))
}

func TestGetMsg_EquipmentNotFound(t *testing.T) {
	assert.Equal(t, "设备不存在", errcode.GetMsg(errcode.ErrEquipmentNotFound))
}

func TestGetMsg_StockInsufficient(t *testing.T) {
	assert.Equal(t, "设备库存不足", errcode.GetMsg(errcode.ErrStockInsufficient))
}

func TestGetMsg_DuplicateBorrow(t *testing.T) {
	assert.Equal(t, "已有该设备的未归还借阅", errcode.GetMsg(errcode.ErrDuplicateBorrow))
}

func TestGetMsg_BorrowApproveFailed(t *testing.T) {
	assert.Equal(t, "审批失败，库存不足", errcode.GetMsg(errcode.ErrBorrowApproveFailed))
}

func TestGetMsg_BorrowStatusInvalid(t *testing.T) {
	assert.Equal(t, "工单状态异常", errcode.GetMsg(errcode.ErrBorrowStatusInvalid))
}

func TestGetMsg_InternalServer(t *testing.T) {
	assert.Equal(t, "系统内部错误", errcode.GetMsg(errcode.ErrInternalServer))
}

func TestGetMsg_UnknownCode(t *testing.T) {
	assert.Equal(t, "未知错误", errcode.GetMsg(9999))
}

func TestErrorCodes_NoDuplicates(t *testing.T) {
	codes := []int{
		errcode.Success,
		errcode.ErrInvalidParam,
		errcode.ErrNotFound,
		errcode.ErrConflict,
		errcode.ErrInternal,
		errcode.ErrAuthFailed,
		errcode.ErrAccountDisabled,
		errcode.ErrTokenMissing,
		errcode.ErrTokenInvalid,
		errcode.ErrPermissionDenied,
		errcode.ErrUserExists,
		errcode.ErrUserNotFound,
		errcode.ErrEquipmentNotFound,
		errcode.ErrStockInsufficient,
		errcode.ErrDuplicateBorrow,
		errcode.ErrBorrowApproveFailed,
		errcode.ErrBorrowStatusInvalid,
		errcode.ErrInternalServer,
	}
	seen := make(map[int]bool)
	for _, c := range codes {
		assert.False(t, seen[c], "error code %d is duplicated", c)
		seen[c] = true
	}
}

func TestErrorCodes_Segments(t *testing.T) {
	// 1xxx: 通用错误
	assert.True(t, errcode.ErrInvalidParam >= 1000 && errcode.ErrInvalidParam < 2000)
	assert.True(t, errcode.ErrNotFound >= 1000 && errcode.ErrNotFound < 2000)
	// 2xxx: 认证与权限
	assert.True(t, errcode.ErrAuthFailed >= 2000 && errcode.ErrAuthFailed < 3000)
	assert.True(t, errcode.ErrPermissionDenied >= 2000 && errcode.ErrPermissionDenied < 3000)
	// 3xxx: 业务错误
	assert.True(t, errcode.ErrUserExists >= 3000 && errcode.ErrUserExists < 4000)
	assert.True(t, errcode.ErrBorrowStatusInvalid >= 3000 && errcode.ErrBorrowStatusInvalid < 4000)
	// 5xxx: 系统错误
	assert.True(t, errcode.ErrInternalServer >= 5000 && errcode.ErrInternalServer < 6000)
}
