package messagequeue

import (
	"github.com/yangkushu/rum-go/iface"
)

type KafkaConfig struct {
	Brokers    string `mapstructure:"brokers" yaml:"brokers"`
	Username   string `mapstructure:"username" yaml:"username"`
	Password   string `mapstructure:"password" yaml:"password"`
	CaFile     string `mapstructure:"ca_file" yaml:"ca_file"`
	Mechanisms string `mapstructure:"mechanisms" yaml:"mechanisms"`
	Protocol   string `mapstructure:"protocol" yaml:"protocol"`
	IsDebug    bool   `mapstructure:"is_debug" yaml:"is_debug"`
	Logger     iface.ILogger
}
