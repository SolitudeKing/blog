package pagination

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ListPage 是增长型列表分页信息，固定放在响应体的 data 之外。
// count 是当前页条目数（= len(data)），不是 total，避免大表 COUNT 拖慢列表接口。
type ListPage struct {
	Page     int  `json:"page"`      // 当前页码 1-based
	PageSize int  `json:"page_size"` // 每页大小
	Count    int  `json:"count"`     // 当前页条目数
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
