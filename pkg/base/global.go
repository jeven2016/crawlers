package base

import "context"

var (
	internalCfg *ServerConfig
	globalCtx   context.Context
)

func SetConfig(cfg *ServerConfig) {
	internalCfg = cfg
}

func GetConfig() *ServerConfig {
	return internalCfg
}

func SetSystemContext(ctx context.Context) {
	globalCtx = ctx
}

func GetSystemContext() context.Context {
	return globalCtx
}
