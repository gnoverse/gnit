package config

type Config struct {
	RealmPath string
	Remote    string
	ChainID   string
	GasFee    string
	GasWanted string
	Account   string
}

func DefaultConfig() *Config {
	return &Config{
		RealmPath: "gno.land/r/example",
		Remote:    "tcp://127.0.0.1:26657",
		ChainID:   "dev",
		GasFee:    "10000000ugnot",
		GasWanted: "5000000000",
		Account:   "test",
	}
}
