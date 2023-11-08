package base

var (
	internalCfg *ServerConfig
)

func SetConfig(cfg *ServerConfig) {
	internalCfg = cfg
}

func GetConfig() *ServerConfig {
	return internalCfg
}
