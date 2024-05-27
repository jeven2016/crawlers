package base

import "fmt"

type ApiResult struct {
	*AppError
	Ok          bool              `json:"ok"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
	Payload     any               `json:"payload,omitempty"`
}

func Fails(code ErrorCode) *ApiResult {
	return &ApiResult{
		Ok: false,
		AppError: &AppError{
			Code:    code,
			Message: GetErrMessage(code),
		},
	}
}

func FailsWithMessage(code ErrorCode, message string) *ApiResult {
	return &ApiResult{
		Ok: false,
		AppError: &AppError{
			Code:    code,
			Message: message,
		},
	}
}

func FailsWithParams(code ErrorCode, params ...string) *ApiResult {
	return &ApiResult{
		Ok: false,
		AppError: &AppError{
			Code:    code,
			Message: GetErrMessage(code, params),
		},
	}
}

func Success(payload any) *ApiResult {
	return &ApiResult{
		Ok: true,
		AppError: &AppError{
			Code: ErrCodeOK,
		},
		Payload: payload,
	}
}

func SuccessCode(code ErrorCode, params ...string) *ApiResult {
	return &ApiResult{
		Ok: true,
		AppError: &AppError{
			Code:    code,
			Message: GetErrMessage(code, params),
		},
	}
}

func GetErrMessage(errCode ErrorCode, params ...any) string {
	if val, ok := errMap[errCode]; ok {
		return fmt.Sprintf(val, params...)
	}
	return errMap[ErrCodeUnknown]

}
