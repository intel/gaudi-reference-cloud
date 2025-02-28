package config

type Config struct {
	SchedulerInterval uint16 `koanf:"schedulerInterval"`
	IDCServiceConfig  struct {
		StorageAPIGrpcEndpoint string `koanf:"storageApiServerAddr"`
	} `koanf:"idcServiceConfig"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		return &Config{SchedulerInterval: 2}
	}
	return Cfg
}
