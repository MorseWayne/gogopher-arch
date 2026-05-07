package config

import "os"

type Config struct {
	Port       string
	SandboxURL string
	DB_URL     string
	RedisUrl   string
}

func Load() Config {
	return Config{
		Port:       envOrDefault("PORT", ":8080"),
		SandboxURL: envOrDefault("SANDBOX_URL", "http://localhost:8081/execute"),
		DB_URL:     envOrDefault("DB_URL", "postgres://user:pass@localhost:5432/gogopher?sslmode=disable"),
		RedisUrl:   envOrDefault("REDIS_URL", "localhost:6379"),
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}