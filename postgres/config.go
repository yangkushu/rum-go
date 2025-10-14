package postgres

import (
	"fmt"
)

type Config struct {
	Host            string `mapstructure:"host" yaml:"host"`                           // 数据库服务器地址
	Port            string `mapstructure:"port" yaml:"port"`                           // 数据库服务器端口
	User            string `mapstructure:"user" yaml:"user"`                           // 数据库用户
	Password        string `mapstructure:"password" yaml:"password"`                   // 数据库密码
	DBName          string `mapstructure:"dbname" yaml:"dbname"`                       // 数据库名称
	SSLMode         string `mapstructure:"ssl_mode" yaml:"ssl_mode"`                   // SSL模式
	ConnectTimeout  int    `mapstructure:"connect_time_out" yaml:"connect_time_out"`   // 连接超时设置 单位秒
	TimeZone        string `mapstructure:"timezone" yaml:"timezone"`                   // 服务器时区
	MaxIdleConns    int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`       // 连接池中的最大空闲连接数
	MaxOpenConns    int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`       // 最大打开的连接数
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"` // 连接的最大可复用时间 单位分钟
	LogLevel        string `mapstructure:"log_level" yaml:"log_level"`                 // 日志级别  silent error  warn info
	DefaultSchema   string `mapstructure:"default_schema" yaml:"default_schema"`       // 默认schema
	DryRun          bool   `mapstructure:"dry_run" yaml:"dry_run"`                     // // DryRun generate sql without execute
}

func (c *Config) ToDSN() (string, error) {

	str := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName)

	if len(c.SSLMode) > 0 {
		str += fmt.Sprintf(" sslmode=%s", c.SSLMode)
	}
	if c.ConnectTimeout > 0 {
		str += fmt.Sprintf(" connect_timeout=%d", c.ConnectTimeout)
	}
	if len(c.TimeZone) > 0 {
		str += fmt.Sprintf(" timezone=%s", c.TimeZone)
	}
	if len(c.DefaultSchema) > 0 {
		str += fmt.Sprintf(" search_path=%s", c.DefaultSchema)
	}
	return str, nil
}
