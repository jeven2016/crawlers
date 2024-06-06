package base

// ErrorCode Error represents an error returned from the server
var ErrorCode *errorCodeInfo

// errorCodeInfo error code information
type errorCodeInfo struct {
	OK int

	NotFound   int
	Unexpected int
	BadRequest int
	Required   int
	Duplicated int

	SiteNotFound          int
	ProcessorNotFound     int
	IllegalPageUrl        int
	ExcludedNovelPageTask int
	IdsRequired           int
}

func init() {
	ErrorCode = &errorCodeInfo{
		OK:         200,
		NotFound:   404,
		Unexpected: 500,
		BadRequest: 400,

		Required:   1003,
		Duplicated: 1004,

		SiteNotFound:          1100,
		ProcessorNotFound:     1101,
		IllegalPageUrl:        1102,
		ExcludedNovelPageTask: 1103,
		IdsRequired:           1104,
	}
}
