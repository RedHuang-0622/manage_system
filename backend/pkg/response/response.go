package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Msg: "创建成功", Data: data})
}

func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{Code: code, Msg: msg, Data: nil})
}

// PageRequest 分页请求基类
type PageRequest struct {
	Page     int `form:"page"      json:"page"      binding:"omitempty,gte=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"omitempty,gte=1,lte=100"`
}

// PageResult 分页响应体
type PageResult struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     interface{} `json:"list"`
}

// Normalize 设置分页默认值并返回 offset, limit
func (p *PageRequest) Normalize() (offset, limit int) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	limit = p.PageSize
	if limit > 100 {
		limit = 100
	}
	return (p.Page - 1) * limit, limit
}

// NormalizePagination 统一分页参数规范化，返回 (page, pageSize, offset)。
// defaultSize: 未指定时的默认 page_size；maxSize: 上限（超过则截断）。
func NormalizePagination(page, pageSize, defaultSize, maxSize int) (int, int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	offset := (page - 1) * pageSize
	return page, pageSize, offset
}

func PageResponse(c *gin.Context, total int64, page, pageSize int, list interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: PageResult{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     list,
		},
	})
}
