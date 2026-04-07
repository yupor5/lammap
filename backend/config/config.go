package config

import "os"

type Config struct {
	Port         string
	DBPath       string
	JWTSecret    string
	OpenAIAPIKey string
	OpenAIModel  string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DBPath:       getEnv("DB_PATH", "quotepro.db"),
		JWTSecret:    getEnv("JWT_SECRET", "quotepro-secret-key-change-in-production"),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4o-mini"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
