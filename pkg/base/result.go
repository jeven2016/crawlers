package base

type ApiResult struct {
	Ok          bool              `json:"ok"`
	Code        ErrorCode         `json:"code,omitempty"`
	Message     string            `json:"message,omitempty"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
	Payload     any               `json:"payload,omitempty"`
}

func Fails(code ErrorCode) *ApiResult {
	return &ApiResult{
		Ok:      false,
		Code:    code,
		Message: GetErrMessage(code),
	}
}

func FailsWithMessage(code ErrorCode, message string) *ApiResult {
	return &ApiResult{
		Ok:      false,
		Code:    code,
		Message: message,
	}
}

func FailsWithParams(code ErrorCode, params ...string) *ApiResult {
	return &ApiResult{
		Ok:      false,
		Code:    code,
		Message: GetErrMessage(code, params),
	}
}

func Success(payload any) *ApiResult {
	return &ApiResult{
		Ok:      true,
		Code:    ErrCodeOK,
		Payload: payload,
	}
}

func SuccessCode(code ErrorCode, params ...string) *ApiResult {
	return &ApiResult{
		Ok:      true,
		Code:    code,
		Message: GetErrMessage(code, params),
	}
}
