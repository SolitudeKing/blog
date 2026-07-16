package errors

import "net/http"

type AppError struct {
	Code       int
	Message    string
	HTTPStatus int
}

func (e AppError) Error() string {
	return e.Message
}

func New(code int) AppError {
	return AppError{
		Code:       code,
		Message:    Message(code),
		HTTPStatus: HTTPStatus(code),
	}
}

func HTTPStatus(code int) int {
	switch code {
	case CodeUnauthorized, CodeTokenExpired, CodeInvalidToken, CodeRefreshTokenExpired:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeResourceNotFound:
		return http.StatusNotFound
	case CodeResourceConflict, CodeDuplicateSlug, CodeInvalidResourceState:
		return http.StatusConflict
	case CodeRateLimited, CodeOperationTooFrequent:
		return http.StatusTooManyRequests
	case CodeUploadFileTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeUnsupportedFileType:
		return http.StatusUnsupportedMediaType
	case CodeDatabaseUnavailable, CodeCacheUnavailable, CodeStorageUnavailable:
		return http.StatusServiceUnavailable
	case CodeInternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
