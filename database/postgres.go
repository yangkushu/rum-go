package database

import (
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strings"
	"time"
)

func NewPostgres(c *PostgresConfig) (*gorm.DB, error) {

	dsn, err := c.ToDSN()

	if err != nil {
		return nil, fmt.Errorf("failed to decrypt postgres password: %w", err)
	}

	nc := &gorm.Config{
		DryRun: c.DryRun,
	}

	setLogger(c, nc)

	db, err := gorm.Open(postgres.Open(dsn), nc)
	if err != nil {
		return nil, errors.Wrap(err, "new postgres error")
	}

	// 这样设置只能修改一个连接，不能修改连接池中的连接
	//设置默认 schema
	//if len(c.DefaultSchema) > 0 {
	//	db.Exec(fmt.Sprintf("SET search_path TO %s", c.DefaultSchema))
	//}

	// 注册回调 当前版本先去掉
	//err = db.Callback().Query().Before("gorm:query").Register("autoSelectFields", autoSelectFields)
	//if err != nil {
	//	return nil, err
	//}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Minute * time.Duration(c.ConnMaxLifetime))

	return db, nil
}

func NewPostgresWithCallback(c *PostgresConfig, callback func(db *gorm.DB) error) (*gorm.DB, error) {
	db, err := NewPostgres(c)
	if err != nil {
		return nil, errors.Wrap(err, "new postgres error")
	}

	err = callback(db)
	if err != nil {
		return nil, errors.Wrap(err, "set callback error")
	}

	return db, nil
}

func setLogger(c *PostgresConfig, gc *gorm.Config) {

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
		//gc.Logger = logger.Default.LogMode(logLevel)
		//}

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
