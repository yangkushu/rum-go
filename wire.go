package rum

import (
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/yangkushu/rum-go/config"
	"github.com/yangkushu/rum-go/log"
	"github.com/yangkushu/rum-go/postgres"
)

func ProvideConfig() (*Config, error) {
	loader := config.NewConfigLoader()
	cfg := &Config{}
	if err := loader.Load(cfg); err != nil {
		return nil, errors.Wrap(err, "init config loader failed")
	}
	return cfg, nil
}

// MinimalSet 提供最小依赖配置（仅日志和配置）
var MinimalSet = wire.NewSet(
	ProvideConfig,
	wire.FieldsOf(new(*Config), "log"),
	log.NewLogger,
)

// PostgresSet 提供数据库相关依赖（默认无选项）
var PostgresSet = wire.NewSet(
	wire.FieldsOf(new(*Config), "postgres"),
	postgres.NewPostgres,
	ProvideDefaultPostgresOptions,
)

// PostgresWithOptionSet 提供数据库相关依赖（需要自定义选项）
var PostgresWithOptionSet = wire.NewSet(
	wire.FieldsOf(new(*Config), "postgres"),
	postgres.NewPostgres,
)

// ProvideDefaultPostgresOptions 提供默认的空选项
func ProvideDefaultPostgresOptions() []postgres.Option {
	return []postgres.Option{}
}
