package config

import (
	"encoding/json"
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

type GnitFile struct {
	RealmPath   string   `json:"realm_path"`
	StagedFiles []string `json:"staged_files"`
}

func DefaultConfig() *Config {
	realmPath := "gno.land/r/example"

	if data, err := os.ReadFile(".gnit"); err == nil {
		content := strings.TrimSpace(string(data))

		var gnitFile GnitFile
		if err := json.Unmarshal([]byte(content), &gnitFile); err == nil {
			if gnitFile.RealmPath != "" {
				realmPath = gnitFile.RealmPath
			}
		} else if content != "" {
			realmPath = content
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
