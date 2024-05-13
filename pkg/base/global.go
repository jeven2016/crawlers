package base

import (
	"context"
)

var (
	globalCtx context.Context
)

func SetSystemContext(ctx context.Context) {
	globalCtx = ctx
}

func GetSystemContext() context.Context {
	return globalCtx
}
