package database

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 這裡暫時將dsn設為固定，之後需要放env再改
func Connect() (*gorm.DB, error) {

	dsn := fmt.Sprintf(
		"postgres://%s:%s@amdb.c96ia8gyqaxe.ap-southeast-2.rds.amazonaws.com:5432/postgres?sslmode=verify-full&sslrootcert=./global-bundle.pem",
		os.Getenv("DB_UISER"),
		os.Getenv("DB_PASSWORD"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	// 連線池設定
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// 確認可以連線到 PostgreSQL
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
