package pagination

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type CursorPage struct {
	Cursor     string `json:"cursor"`
	NextCursor string `json:"next_cursor"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"has_more"`
}

func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}
