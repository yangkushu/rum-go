package log

type Config struct {
	Level             string `mapstructure:"level" yaml:"level"`
	Development       bool   `mapstructure:"development" yaml:"development"`
	DisableCaller     bool   `mapstructure:"disable_caller" yaml:"disable_caller"`
	CallerSkip        int    `mapstructure:"caller_skip" yaml:"caller_skip"`
	DisableStacktrace bool   `mapstructure:"disable_stacktrace" yaml:"disable_stacktrace"`
	Encoding          string `mapstructure:"encoding" yaml:"encoding"`
	TimeFormat        string `mapstructure:"time_format" yaml:"time_format"`

	EnableWriteToMemory bool `mapstructure:"enable_write_to_memory" yaml:"enable_write_to_memory"` // 开启内存写入同步
	MemoryMaxMB         int  `mapstructure:"memory_max_mb" yaml:"memory_max_mb"`                   // 内存日志最大占用

	EnableWriteToFile bool   `mapstructure:"enable_write_to_file" yaml:"enable_write_to_file"` // 开启文件写入同步
	LogFile           string `mapstructure:"log_file" yaml:"log_file"`                         // 输出到文件

	RollingFile *RollingFileConfig `mapstructure:"rolling_file" yaml:"rolling_file"`

	WriteSyncerChan     chan<- []byte
	WriteSyncerEncoding string `mapstructure:"write_syncer_encoding" yaml:"write_syncer_encoding"`
	WriteSyncerLevel    string `mapstructure:"write_syncer_level" yaml:"write_syncer_level"`
}

type RollingFileConfig struct {
	MaxSize    int  `mapstructure:"max_size" yaml:"max_size"` // megabytes
	MaxBackups int  `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAge     int  `mapstructure:"max_age" yaml:"max_age"`       //days
	LocalTime  bool `mapstructure:"local_time" yaml:"local_time"` // disabled by default
	Compress   bool `mapstructure:"compress" yaml:"compress"`     // disabled by default
}
