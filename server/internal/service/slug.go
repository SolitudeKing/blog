package service

import (
	"regexp"

	apperrors "solitude-blog/server/internal/errors"
)

// slugMaxLength 与 model.Tag / model.Topic 的 `gorm:"size:120"` 保持一致。
// 文章模型列宽为 size:220 仅作存量兼容，对外校验契约统一为 120。
const slugMaxLength = 120

// slugFormatPattern 定义公开 slug 的稳定契约：小写字母、数字与连字符分段，
// 不允许大写、下划线、空格、中文或首尾连字符。
var slugFormatPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// validateSlugFormat 校验 slug 格式与长度，不合法返回 CodeInvalidParameter（HTTP 400）。
// 正则保证纯 ASCII，len(slug) 的字节数即字符数。空串由各服务的必填校验先行处理。
func validateSlugFormat(slug string) error {
	if !slugFormatPattern.MatchString(slug) || len(slug) > slugMaxLength {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}
	return nil
}
