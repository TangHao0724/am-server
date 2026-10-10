package database

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 這裡暫時將dsn設為固定，之後需要放env再改
func Connect() (*gorm.DB, error) {

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")

	// 避免環境變數缺漏時產生難以判讀的連線錯誤
	if user == "" || password == "" || host == "" {
		return nil, fmt.Errorf("missing required database environment variables")
	}

	u := &url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(host, "5432"),
		Path:   "postgres",
		User:   url.UserPassword(user, password),
	}

	q := u.Query()
	q.Set("sslmode", "verify-full")
	q.Set("sslrootcert", "/home/ubuntu/am-server/global-bundle.pem")
	u.RawQuery = q.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{})
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
