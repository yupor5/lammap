package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port         string
	DBPath       string
	JWTSecret    string
	OpenAIAPIKey string
	OpenAIModel  string
	OpenAIBase   string
}

func Load() *Config {
	// 读取可选的配置文件（默认 ./config.yaml）。环境变量始终优先生效。
	// 这样本地可以把密钥放在 config.yaml，中/线上再用环境变量覆盖。
	type fileCfg struct {
		Port      string `mapstructure:"port"`
		DBPath    string `mapstructure:"dbPath"`
		JWTSecret string `mapstructure:"jwtSecret"`
		AI        struct {
			BaseURL string `mapstructure:"baseUrl"`
			APIKey  string `mapstructure:"apiKey"`
			Model   string `mapstructure:"model"`
		} `mapstructure:"ai"`
	}

	fc := fileCfg{}
	cfgPath := strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	v := viper.New()
	if cfgPath != "" {
		v.SetConfigFile(cfgPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		// 支持从 backend/ 目录启动或从仓库根目录启动
		v.AddConfigPath(".")
		v.AddConfigPath("..")
		v.AddConfigPath("./backend")
	}
	if err := v.ReadInConfig(); err == nil {
		_ = v.Unmarshal(&fc)
	}

	return &Config{
		Port:      getEnv("PORT", firstNonEmpty(fc.Port, "8080")),
		DBPath:    getEnv("DB_PATH", firstNonEmpty(fc.DBPath, "quotepro.db")),
		JWTSecret: getEnv("JWT_SECRET", firstNonEmpty(fc.JWTSecret, "quotepro-secret-key-change-in-production")),

		// OpenAI 兼容配置：用于 Ark / OpenAI 等 chat/completions 网关
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", fc.AI.APIKey),
		OpenAIModel:  getEnv("OPENAI_MODEL", firstNonEmpty(fc.AI.Model, "gpt-4o-mini")),
		OpenAIBase:   getEnv("OPENAI_BASE_URL", firstNonEmpty(fc.AI.BaseURL, "https://api.openai.com/v1")),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
