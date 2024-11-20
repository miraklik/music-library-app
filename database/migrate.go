package database

import (
	"log"
	"music-library/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&models.Song{})
	if err != nil {
		log.Fatal("Migration failed: ", err)
	}
}
