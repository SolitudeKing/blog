package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ListPage 是增长型列表分页信息，固定放在响应体的 data 之外。
// count 是当前页条目数（= len(data)）；total 是符合条件的总条数，
// 用于计算总页数与"跳到第 N 页"等场景。
type ListPage struct {
	Page     int  `json:"page"`      // 当前页码 1-based
	PageSize int  `json:"page_size"` // 每页大小
	Count    int  `json:"count"`     // 当前页条目数
	Total    int  `json:"total"`     // 符合条件的总条数
	HasMore  bool `json:"has_more"`  // 是否还有下一页
}

// NormalizePage 把客户端传入的 page 钳到合法范围。
// <= 0 视为未传，回落到第 1 页。
func NormalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

// NormalizePageSize 把客户端传入的 page_size 钳到 [DefaultPageSize, MaxPageSize]。
// <= 0 视为未传，使用默认值。
func NormalizePageSize(size int) int {
	if size <= 0 {
		return DefaultPageSize
	}
	if size > MaxPageSize {
		return MaxPageSize
	}
	return size
}

// BindPage 从 gin context 读取 page / page_size query 参数并规范化。
// 用于消除 6 个 handler 中重复的 strconv.Atoi(c.Query("page")) 模式。
// 错误（如非数字）会被钳到默认值而不是报错——分页参数非法时回退首页更友好。
func BindPage(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.Query("page"))
	pageSize, _ = strconv.Atoi(c.Query("page_size"))
	return NormalizePage(page), NormalizePageSize(pageSize)
}
