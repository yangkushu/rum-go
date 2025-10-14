package nacos

type Config struct {
	Addr            string `mapstructure:"addr" yaml:"addr"`
	Port            uint64 `mapstructure:"port" yaml:"port"`
	Path            string `mapstructure:"path" yaml:"path"`
	Namespace       string `mapstructure:"namespace" yaml:"namespace"`
	LogLevel        string `mapstructure:"loglevel" yaml:"loglevel"`
	Timeout         uint64 `mapstructure:"timeout" yaml:"timeout"`
	BeatInterval    int64  `mapstructure:"beat_interval" yaml:"beat_interval"`
	UpdateThreadNum int    `mapstructure:"update_thread_num" yaml:"update_thread_num"`
	CacheDir        string `mapstructure:"cache_dir" yaml:"cache_dir"`
	LogDir          string `mapstructure:"log_dir" yaml:"log_dir"`
	Service         string `mapstructure:"service" yaml:"service"`
	Group           string `mapstructure:"group" yaml:"group"`
	Cluster         string `mapstructure:"cluster" yaml:"cluster"`
	ListenPort      uint64 `mapstructure:"listen_port" yaml:"listen_port"`
	DataId          string `mapstructure:"data_id" yaml:"data_id"`
	Scheme          string `mapstructure:"scheme" yaml:"scheme"`
	Username        string `mapstructure:"username" yaml:"username"`
	Password        string `mapstructure:"password" yaml:"password"`
}
