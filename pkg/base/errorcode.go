package base

import "fmt"

type ErrorCode int

const (
	ErrCodeUnknown ErrorCode = 0
	ErrCodeOK      ErrorCode = 200

	NotFound ErrorCode = 1000 + iota
	ErrCodeUnSupportedCatalog
	ErrCodePageUrlRequired
	ErrCodeTaskSubmitted
	ErrPublishMessage
	ErrParsePageUrl
	ErrDuplicated
	ErrCatalogNotFound
	ErrSiteNotFound
	ErrExcludedNovel
)

var errMap = map[ErrorCode]string{}

func init() {
	errMap[ErrCodeUnknown] = "unexpected error occurred"
	errMap[NotFound] = "not found"
	errMap[ErrCodeUnSupportedCatalog] = "unsupported catalog '%s'"
	errMap[ErrCodePageUrlRequired] = "pageUrl is required"
	errMap[ErrCodeTaskSubmitted] = "a task is already submitted"
	errMap[ErrPublishMessage] = "failed to publish the message, reason: %s"
	errMap[ErrParsePageUrl] = "failed to parse the page testUrl, reason: %s"
	errMap[ErrDuplicated] = "it's duplicated to save with %v(%v)"
	errMap[ErrCatalogNotFound] = "catalog '%s' not found"
	errMap[ErrSiteNotFound] = "site '%s' not found"
	errMap[ErrExcludedNovel] = "excluded novel task submitted"
}

func GetErrMessage(errCode ErrorCode, params ...any) string {
	if val, ok := errMap[errCode]; ok {
		return fmt.Sprintf(val, params...)
	}
	return errMap[ErrCodeUnknown]

}
