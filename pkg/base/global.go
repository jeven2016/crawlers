package base

import (
	"github.com/jeven2016/mylibs/system"
)

var (
	internalCfg *ServerConfig
	sys         *system.System
)

func SetConfig(cfg *ServerConfig) {
	internalCfg = cfg
}

func GetConfig() *ServerConfig {
	return internalCfg
}

func GetSystem() *system.System {
	return sys
}

func SetSystem(s *system.System) {
	sys = s
}
