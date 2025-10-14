package commands

import (
	"fmt"
	"os"

	"gnit/config"
	"gnit/gnokey"
)

type Clone struct {
	client *gnokey.Client
	config *config.Config
}

func NewClone(client *gnokey.Client, cfg *config.Config) *Clone {
	return &Clone{
		client: client,
		config: cfg,
	}
}

func (c *Clone) Execute(realmPath string) error {
	fmt.Printf("Cloning repository from '%s'...\n", realmPath)

	if _, err := os.Stat(".gnit"); err == nil {
		return fmt.Errorf(".gnit file already exists in current directory")
	}

	if err := os.WriteFile(".gnit", []byte(realmPath+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create .gnit file: %w", err)
	}

	fmt.Printf("Initialized gnit repository with realm: %s\n", realmPath)

	cloneCfg := *c.config
	cloneCfg.RealmPath = realmPath

	pull := NewPull(c.client, &cloneCfg)
	if err := pull.ExecuteAll(); err != nil {
		return fmt.Errorf("failed to pull files: %w", err)
	}

	fmt.Println("\nRepository cloned successfully!")
	return nil
}
