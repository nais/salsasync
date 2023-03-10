package config

type Config struct {
	MetricsBindAddress string `json:"bind-address"`
	LogLevel           string `json:"log-level"`
	StorageApi         string `json:"storage-api"`
	StorageApiKey      string `json:"storage-api-key"`
	ConsoleApiKey      string `json:"console-api-key"`
}

func DefaultConfig() *Config {
	return &Config{
		MetricsBindAddress: "127.0.0.1:8080",
		LogLevel:           "info",
	}
}
