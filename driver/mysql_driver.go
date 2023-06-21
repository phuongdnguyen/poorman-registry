package driver

import (
	"fmt"

	"github.com/nduyphuong/reverse-registry/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQLDB(host string, user string, password string, dbName string) (*gorm.DB, error) {
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local", user, password, host, dbName)
	db, err := gorm.Open(mysql.Open(dataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(
		&model.ImageModel{},
	)
	return db, nil
}
