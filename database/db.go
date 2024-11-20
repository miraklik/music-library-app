package database

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db   *gorm.DB
	once sync.Once
}

type DatabaseInterface interface {
	Connect() error
	GetDB() *gorm.DB
	Close() error
}

func NewDatabase() *Database {
	return &Database{}
}

func (d *Database) Connect() error {
	var err error

	d.once.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			err = fmt.Errorf("environment variable DATABASE_URL is not set")
			return
		}

		d.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("Error connecting to database: %v\n", err)
			return
		}

		sqlDB, err := d.db.DB()
		if err != nil {
			log.Printf("Error getting sql.DB from gorm.DB: %v\n", err)
			return
		}

		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)

		log.Println("Successfully connected to the database!")
	})

	return err
}

func (d *Database) GetDB() *gorm.DB {
	if d.db == nil {
		log.Fatal("Database connection is not initialized. Call Connect() first.")
	}
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
