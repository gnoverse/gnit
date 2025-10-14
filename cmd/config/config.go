package config

import (
	"os"
	"strings"
)

type Config struct {
	RealmPath string
	Remote    string
	ChainID   string
	GasFee    string
	GasWanted string
	Account   string
}

func DefaultConfig() *Config {
	realmPath := "gno.land/r/example"

	if data, err := os.ReadFile(".gnit"); err == nil {
		if path := strings.TrimSpace(string(data)); path != "" {
			realmPath = path
		}
	}

	return &Config{
		RealmPath: realmPath,
		Remote:    "tcp://127.0.0.1:26657",
		ChainID:   "dev",
		GasFee:    "10000000ugnot",
		GasWanted: "5000000000",
		Account:   "test",
	}
}
