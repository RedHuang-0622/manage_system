package errcode

const (
	Success = 0

	// 通用错误 (1xxx)
	ErrInvalidParam = 1001
	ErrNotFound     = 1002
	ErrConflict     = 1003
	ErrInternal     = 1004

	// 认证与权限 (2xxx)
	ErrAuthFailed       = 2001
	ErrAccountDisabled  = 2002
	ErrTokenMissing     = 2003
	ErrTokenInvalid     = 2004
	ErrPermissionDenied = 2005
	ErrRateLimited      = 2006

	// 业务错误 (3xxx)
	ErrUserExists          = 3001
	ErrUserNotFound        = 3002
	ErrEquipmentNotFound   = 3003
	ErrStockInsufficient   = 3004
	ErrDuplicateBorrow     = 3005
	ErrBorrowApproveFailed = 3006
	ErrBorrowStatusInvalid = 3007

	// 系统错误 (5xxx)
	ErrInternalServer = 5000
)

var codeMsg = map[int]string{
	Success:               "success",
	ErrInvalidParam:       "请求参数错误",
	ErrNotFound:           "资源不存在",
	ErrConflict:           "资源冲突",
	ErrInternal:           "内部错误",
	ErrAuthFailed:         "用户名或密码错误",
	ErrAccountDisabled:    "账号已被禁用",
	ErrTokenMissing:       "未提供认证Token",
	ErrTokenInvalid:       "Token无效或已过期",
	ErrPermissionDenied:   "权限不足",
	ErrRateLimited:        "请求过于频繁",
	ErrUserExists:         "用户名已存在",
	ErrUserNotFound:       "用户不存在",
	ErrEquipmentNotFound:  "设备不存在",
	ErrStockInsufficient:  "设备库存不足",
	ErrDuplicateBorrow:    "已有该设备的未归还借阅",
	ErrBorrowApproveFailed: "审批失败，库存不足",
	ErrBorrowStatusInvalid: "工单状态异常",
	ErrInternalServer:     "系统内部错误",
}

func GetMsg(code int) string {
	if msg, ok := codeMsg[code]; ok {
		return msg
	}
	return "未知错误"
}
