package database

import (
	"fmt"

	"github.com/PrathamDev/storyapp/config"
	"github.com/PrathamDev/storyapp/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.FetchKey("DB_USER"), config.FetchKey("DB_PASSWORD"), config.FetchKey("DB_HOST"), config.FetchKey("DB_PORT"), config.FetchKey("DB_NAME"))
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect to database")
	}

	fmt.Println("Connection opened to database")
	DB.AutoMigrate(&model.User{})
	fmt.Println("Database Migrated")
}
