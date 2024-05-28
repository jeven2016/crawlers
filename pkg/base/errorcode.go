package base

// ErrorCode Error represents an error returned from the server
var ErrorCode *errorCodeInfo

// errorCodeInfo error code information
type errorCodeInfo struct {
	Unknown int
	OK      int

	NotFound   int
	Unexpected int
	Required   int
	Duplicated int
}

func init() {
	ErrorCode = &errorCodeInfo{
		OK:      100,
		Unknown: 101,

		NotFound:   1000,
		Unexpected: 1002,
		Required:   1003,
		Duplicated: 1004,
	}
}
