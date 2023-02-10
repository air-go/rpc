package gorm

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMemoryDB() *gorm.DB {
	var db *gorm.DB
	var err error
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      false,       // 禁用彩色打印
		},
	)
	dialector := sqlite.Open(":memory:?cache=shared")
	if db, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	}); err != nil {
		panic(err)
	}
	dba, err := db.DB()
	dba.SetMaxOpenConns(1)
	return db
}

func CloseMemoryDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}
