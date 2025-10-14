package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMysql(c Config) (*gorm.DB, error) {
	//dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	//	config.User, config.Password, config.Host, config.Port, config.Db)

	db, err := gorm.Open(mysql.Open(c.ToDSN()), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return db, nil
}
