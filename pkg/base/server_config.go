package base

type ApiServerConfig struct {
	//在golang中，struct tag squash 的作用是将嵌套的结构体中的字段提升到外层结构体中，从而减少结构体的嵌套层数。
	ServerConfig `mapstructure:",squash"`
}

func (c ApiServerConfig) GetServerConfig() *ServerConfig {
	return &c.ServerConfig
}

func (c ApiServerConfig) Validate() error {
	return c.ServerConfig.Validate()
}

func (c ApiServerConfig) Complete() error {
	return c.ServerConfig.Complete()
}
