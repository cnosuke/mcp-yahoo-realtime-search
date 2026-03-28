package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config - Application configuration
type Config struct {
	Log      string `koanf:"log"`
	LogLevel string `koanf:"log_level"`

	HTTP struct {
		Binding        string   `koanf:"binding"`
		EndpointPath   string   `koanf:"endpoint_path"`
		AuthToken      string   `koanf:"auth_token"`
		AllowedOrigins []string `koanf:"allowed_origins"`
	} `koanf:"http"`

	Search struct {
		UserAgent      string `koanf:"user_agent"`
		RequestTimeout int    `koanf:"request_timeout"`
	} `koanf:"search"`
}

// LoadConfig - Load configuration file
func LoadConfig(path string) (*Config, error) {
	k := koanf.New(".")

	// 1. Default values
	if err := k.Load(confmap.Provider(defaultValues(), "."), nil); err != nil {
		return nil, err
	}

	// 2. YAML file (optional: skip if not found)
	if _, err := os.Stat(path); err == nil {
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return nil, err
		}
	}

	// 3. Environment variable overrides
	envOverrides, err := loadEnvOverrides()
	if err != nil {
		return nil, err
	}
	if len(envOverrides) > 0 {
		if err := k.Load(confmap.Provider(envOverrides, "."), nil); err != nil {
			return nil, err
		}
	}

	// 4. Unmarshal
	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultValues() map[string]interface{} {
	return map[string]interface{}{
		"log_level":              "info",
		"http.binding":           "localhost:8080",
		"http.endpoint_path":     "/mcp",
		"search.user_agent":      "",
		"search.request_timeout": 30,
	}
}

func loadEnvOverrides() (map[string]interface{}, error) {
	envMapping := map[string]string{
		"LOG_PATH":             "log",
		"LOG_LEVEL":            "log_level",
		"HTTP_BINDING":         "http.binding",
		"HTTP_ENDPOINT_PATH":   "http.endpoint_path",
		"HTTP_AUTH_TOKEN":      "http.auth_token",
		"HTTP_ALLOWED_ORIGINS": "http.allowed_origins",
		"YRS_USER_AGENT":       "search.user_agent",
		"YRS_REQUEST_TIMEOUT":  "search.request_timeout",
	}

	overrides := make(map[string]interface{})
	for envKey, koanfKey := range envMapping {
		val, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}
		switch koanfKey {
		case "http.allowed_origins":
			if val != "" {
				overrides[koanfKey] = strings.Split(val, ",")
			}
		case "search.request_timeout":
			if v, err := strconv.Atoi(val); err == nil {
				overrides[koanfKey] = v
			}
		default:
			overrides[koanfKey] = val
		}
	}
	return overrides, nil
}
