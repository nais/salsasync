package config

type Config struct {
	MetricsBindAddress string `json:"bind-address"`
	LogLevel           string `json:"log-level"`
	StorageApi         string `json:"storage-api"`
	StorageApiKey      string `json:"storage-api-key"`
	ConsoleApi         string `json:"console-api"`
	ConsoleApiKey      string `json:"console-api-key"`
}

func DefaultConfig() *Config {
	return &Config{
		MetricsBindAddress: "127.0.0.1:8080",
		LogLevel:           "debug",
		StorageApi:         "http://localhost:9001/api/v1/",
		StorageApiKey:      "todo",
		ConsoleApi:         "http://localhost:3000/query",
		ConsoleApiKey:      "key1",
	}
}
