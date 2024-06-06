package base

import (
	"errors"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"strconv"
)

type ApiResult struct {
	*AppError
	Ok          bool              `json:"ok"`
	FieldErrors map[string]string `json:"fieldErrors,omitempty"`
	Payload     any               `json:"payload,omitempty"`
}

func HttpStatusCode(err error) int {
	var internalErr AppError
	var ok bool
	if ok = errors.As(err, &internalErr); ok {
		if internalErr.Code == ErrorCode.OK {
			return internalErr.Code
		}
	}
	return ErrorCode.Unexpected
}

func Fails(ctx *gin.Context, code int) *ApiResult {
	return &ApiResult{
		Ok: false,
		AppError: &AppError{
			Code:    code,
			Message: getErrMessage(ctx, code, nil),
		},
	}
}

func FailsWithMessage(code int, message string) *ApiResult {
	return &ApiResult{
		Ok: false,
		AppError: &AppError{
			Code:    code,
			Message: message,
		},
	}
}

func FailsWithError(c *gin.Context, err error) *ApiResult {
	var internalErr AppError
	var ok bool
	if ok = errors.As(err, &internalErr); ok {

		if internalErr.Message != "" {
			return &ApiResult{
				Ok:       false,
				AppError: &internalErr,
			}
		}

		return &ApiResult{
			Ok: false,
			AppError: &AppError{
				Code:    internalErr.Code,
				Message: getErrMessage(c, internalErr.Code, nil),
			},
		}
	} else {
		return &ApiResult{
			Ok: false,
			AppError: &AppError{
				Code:    ErrorCode.Unexpected,
				Message: err.Error()},
		}
	}
}

func FailsWithParams(ctx *gin.Context, code int, params map[string]string) *ApiResult {
	return &ApiResult{
		Ok: false,
		AppError: &AppError{
			Code:    code,
			Message: getErrMessage(ctx, code, params),
		},
	}
}

func Success(payload any) *ApiResult {
	return &ApiResult{
		Ok: true,
		AppError: &AppError{
			Code: ErrorCode.OK,
		},
		Payload: payload,
	}
}

func SuccessCode(ctx *gin.Context, code int, params map[string]string) *ApiResult {
	return &ApiResult{
		Ok: true,
		AppError: &AppError{
			Code:    code,
			Message: getErrMessage(ctx, code, params),
		},
	}
}

func getErrMessage(ctx *gin.Context, code int, params map[string]string) string {
	var msg string
	if params != nil {
		msg = ginI18n.MustGetMessage(
			ctx,
			&i18n.LocalizeConfig{
				MessageID:    strconv.Itoa(code),
				TemplateData: params,
			})
	} else {
		msg = ginI18n.MustGetMessage(ctx, code)
	}

	return msg
}
