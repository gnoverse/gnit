package client

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

type GnitFile struct {
	StagedFiles []string `json:"staged_files"`
}

func DefaultConfig() *Config {
	realmPath := readPackagePathFromGnomod()
	if realmPath == "" {
		realmPath = "gno.land/r/example"
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

func readPackagePathFromGnomod() string {
	data, err := os.ReadFile("gnomod.toml")
	if err != nil {
		return ""
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkgPath := strings.Trim(parts[1], "\"")
				return pkgPath
			}
		}
	}

	return ""
}
