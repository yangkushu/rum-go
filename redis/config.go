package redis

type Config struct {
	Addrs    string `mapstructure:"addrs" yaml:"addrs"` // Ex:127.0.0.1:7000,127.0.0.1:7001,127.0.0.1:7002
	Addr     string `mapstructure:"addr" yaml:"addr"`   // Deprecated: use Addrs instead
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	DB       int    `mapstructure:"db" yaml:"db"`     // 集群模式不支持配置库
	Port     int    `mapstructure:"port" yaml:"port"` // Deprecated: use Addrs instead
}
