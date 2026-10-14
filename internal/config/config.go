package config

import "os"

type Config struct {
    DbURL string
    Port  string
}

func Load() *Config {
    return &Config{
        // Default values if environment variables aren't set
        DbURL: getEnv("DB_URL", "postgres://user:pass@127.0.0.1:5432/attendance_db?sslmode=disable"),
        Port:  getEnv("PORT", "9000"),
    }
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}