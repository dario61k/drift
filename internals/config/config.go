package config

type Config struct {
	Name    string        `toml:"name"`
	Version string        `toml:"version"`
	Debug   bool          `toml:"debug"`
	Server  ServerConfig  `toml:"server"`
	Logging LoggingConfig `toml:"logging"`
}

type ServerConfig struct {
	Port int `toml:"port"`
}

type LoggingConfig struct {
	Level   string `toml:"level"`
	File    string `toml:"file"`
	Console bool   `toml:"console"`
}

var DefaultConfig = map[string]any{
	"name":    "drift",
	"version": "0.0.1",
	"debug":   false,

	"server.port": 8080,

	"logging.level":   "all",
	"logging.file":    "",
	"logging.console": true,
}
