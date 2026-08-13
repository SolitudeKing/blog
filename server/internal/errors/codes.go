package errors

const (
	CodeOK = 0

	CodeUnauthorized        = 10000
	CodeTokenExpired        = 10001
	CodeInvalidToken        = 10002
	CodeRefreshTokenExpired = 10003
	CodeForbidden           = 10004
	CodeInvalidCredentials  = 10005
	CodeAccountDisabled     = 10006

	CodeInvalidRequest        = 20000
	CodeMissingRequiredField  = 20001
	CodeInvalidParameter      = 20002
	CodeInvalidPage           = 20003
	CodeUnsupportedAPIVersion = 20004
	CodeMalformedJSONBody     = 20005

	CodeResourceNotFound       = 30000
	CodeResourceConflict       = 30001
	CodeDuplicateSlug          = 30002
	CodeInvalidResourceState   = 30003
	CodeReferencedResourceUsed = 30004

	CodeRateLimited          = 40000
	CodeUploadFileTooLarge   = 40001
	CodeUnsupportedFileType  = 40002
	CodeOperationTooFrequent = 40003

	CodeInternalServerError = 50000
	CodeDatabaseUnavailable = 50001
	CodeCacheUnavailable    = 50002
	CodeStorageUnavailable  = 50004
)

var messages = map[int]string{
	CodeOK:                     "ok",
	CodeUnauthorized:           "unauthorized",
	CodeTokenExpired:           "token expired",
	CodeInvalidToken:           "invalid token",
	CodeRefreshTokenExpired:    "refresh token expired",
	CodeForbidden:              "forbidden",
	CodeInvalidCredentials:     "invalid credentials",
	CodeAccountDisabled:        "account disabled",
	CodeInvalidRequest:         "invalid request",
	CodeMissingRequiredField:   "missing required field",
	CodeInvalidParameter:       "invalid parameter",
	CodeInvalidPage:            "invalid page",
	CodeUnsupportedAPIVersion:  "unsupported api version",
	CodeMalformedJSONBody:      "malformed json body",
	CodeResourceNotFound:       "resource not found",
	CodeResourceConflict:       "resource conflict",
	CodeDuplicateSlug:          "duplicate slug",
	CodeInvalidResourceState:   "invalid resource state",
	CodeReferencedResourceUsed: "referenced resource used",
	CodeRateLimited:            "rate limited",
	CodeUploadFileTooLarge:     "upload file too large",
	CodeUnsupportedFileType:    "unsupported file type",
	CodeInternalServerError:    "internal server error",
	CodeDatabaseUnavailable:    "database unavailable",
	CodeCacheUnavailable:       "cache unavailable",
	CodeStorageUnavailable:     "storage unavailable",
}

func Message(code int) string {
	if message, ok := messages[code]; ok {
		return message
	}
	return "unknown error"
}
