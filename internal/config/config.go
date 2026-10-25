package config

import (
	"log"
	"os"
)

type Config struct {
	Name     string       `toml:"name"`
	Version  string       `toml:"version"`
	Download string       `toml:"download"`
	Server   ServerConfig `toml:"server"`
	Stream   StreamConfig `toml:"stream"`
}

type ServerConfig struct {
	Port int `toml:"port"`
}

type StreamConfig struct {
	MaxConnections    int `toml:"max_connections"`
	MaxMsgStreamSize  int `toml:"max_msg_stream_size"`
	MaxFileStreamSize int `toml:"max_file_stream_size"`
}

const maxFileStreamSize = 1024 * 1024 * 1024 * 5

var DefaultConfig = map[string]any{
	"name":     "drift",
	"version":  "0.0.1",
	"download": "",

	"server.port": 4040,

	"stream.max_connections": 32,
	"stream.max_msg_stream_size" : 1000,
	"stream.max_file_stream_size" : maxFileStreamSize,
}

func init() {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	DefaultConfig["download"] = home
}
