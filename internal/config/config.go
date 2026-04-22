package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Config struct {
	ServerName  string
	Host        string
	Port        int
	Party       int
	DBPath      string
	WorkerCount int
}

func Load() Config {
	host := readString("HOST", "0.0.0.0")
	port := readInt("PORT", 8081)
	party := readInt("PARTY", 0)
	dbPath := readString("DB_PATH", filepath.Join("data", defaultDBName(party)))
	workerCount := readInt("WORKER_COUNT", runtime.NumCPU())
	if workerCount < 1 {
		workerCount = 1
	}

	return Config{
		ServerName:  defaultServerName(party),
		Host:        host,
		Port:        port,
		Party:       party,
		DBPath:      dbPath,
		WorkerCount: workerCount,
	}
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func defaultServerName(party int) string {
	if party == 1 {
		return "server-b"
	}
	return "server-a"
}

func defaultDBName(party int) string {
	if party == 1 {
		return "server_b.db"
	}
	return "server_a.db"
}

func readString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func readInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
