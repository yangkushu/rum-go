package postgres

import (
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

const (
	defaultMaxIdleConns    = 10
	defaultMaxOpenConns    = 100
	defaultConnMaxLifetime = 60 // 分钟
)

func NewPostgres(c *Config, options ...Option) (*gorm.DB, error) {

	dsn, err := c.ToDSN()

	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt postgres password")
	}

	nc := &gorm.Config{
		DryRun: c.DryRun,
	}

	setLogger(c, nc)

	db, err := gorm.Open(postgres.Open(dsn), nc)
	if err != nil {
		return nil, errors.Wrap(err, "new postgres error")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "get sql db error")
	}

	// 连接测试
	if err := sqlDB.Ping(); err != nil {
		return nil, errors.Wrap(err, "ping database failed")
	}

	// 连接池设置
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = defaultMaxIdleConns
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = defaultMaxOpenConns
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = defaultConnMaxLifetime
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Minute * time.Duration(c.ConnMaxLifetime))

	if options != nil {
		for _, option := range options {
			if err := option(db); err != nil {
				return nil, errors.Wrap(err, "apply option error")
			}
		}
	}

	return db, nil
}

type Option func(*gorm.DB) error

func WithPlugins(plugins []gorm.Plugin) func(*gorm.DB) error {
	return func(db *gorm.DB) error {
		for _, plugin := range plugins {
			if err := db.Use(plugin); err != nil {
				return errors.Wrap(err, "use plugin error")
			}
		}
		return nil
	}
}

func WithPlugin(plugin gorm.Plugin) func(*gorm.DB) error {
	return func(db *gorm.DB) error {
		if err := db.Use(plugin); err != nil {
			return errors.Wrap(err, "use plugin error")
		}
		return nil
	}
}

func WithCallback(callback func(*gorm.DB) error) Option {
	return func(db *gorm.DB) error {
		return callback(db)
	}
}

func setLogger(c *Config, gc *gorm.Config) {

	var logLevel logger.LogLevel = 0
	switch strings.ToLower(c.LogLevel) {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Info
	}

	if logLevel != 0 {
		gc.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // 慢 SQL 阈值
				LogLevel:                  logLevel,    // 日志级别
				Colorful:                  true,        // 禁用彩色打印
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			},
		)

	}

}
