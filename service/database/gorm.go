package database

import (
	"fmt"
	"url-shortener/helpers"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func NewConnection() *gorm.DB {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		helpers.Getenv("DB_USERNAME", "root"),
		helpers.Getenv("DB_PASSWORD", ""),
		helpers.Getenv("DB_HOST", "127.0.0.1"),
		helpers.Getenv("DB_PORT", "3306"),
		helpers.Getenv("DB_DATABASE", "golang_db"),
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		panic(fmt.Sprintf("can't connect to database: %v", err))
	}

	return db
}
