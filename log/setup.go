package log

type ConfigLog struct {
	Log *struct {
		LogLevel      string `mapstructure:"log_level" yaml:"log_level"`
		LogFile       string `mapstructure:"logfile" yaml:"logfile"`
		Encoding      string `mapstructure:"encoding" yaml:"encoding"`
		DisableCaller bool   `mapstructure:"disable_caller" yaml:"disable_caller"`
		TimeFormat    string `mapstructure:"time_format" yaml:"time_format"`
	} `mapstructure:"log" yaml:"log"`
}

func Init(cfg *ConfigLog) error {
	def := NewDefaultConfig()

	if nil != cfg && nil != cfg.Log {
		switch cfg.Log.LogLevel {
		case "debug", "info", "warn", "error", "dpanic", "panic", "fatal":
			def.Level = cfg.Log.LogLevel
		}

		if len(cfg.Log.Encoding) > 0 {
			switch cfg.Log.Encoding {
			case "console", "json", "yaml", "pro":
				def.Encoding = cfg.Log.Encoding
			}
		}
		def.DisableCaller = cfg.Log.DisableCaller
		def.TimeFormat = cfg.Log.TimeFormat
	}

	if l, err := NewLogger(def); nil == err {
		SetLogger(l)
		return nil
	} else {
		return err
	}
}
