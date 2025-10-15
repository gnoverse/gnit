package commands

import (
	"fmt"
	"os"
	"strings"

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

	repoName := extractRepoName(realmPath)
	if repoName == "" {
		return fmt.Errorf("invalid realm path format")
	}

	if _, err := os.Stat(repoName); err == nil {
		return fmt.Errorf("directory '%s' already exists", repoName)
	}

	if err := os.Mkdir(repoName, 0755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", repoName, err)
	}

	if err := os.Chdir(repoName); err != nil {
		return fmt.Errorf("failed to change to directory '%s': %w", repoName, err)
	}

	gnitFile := config.GnitFile{
		StagedFiles: []string{},
	}

	if err := WriteGnitFileData(&gnitFile); err != nil {
		return fmt.Errorf("failed to create .gnit file: %w", err)
	}

	fmt.Printf("Initialized gnit repository with realm: %s\n", realmPath)

	cloneCfg := *c.config
	cloneCfg.RealmPath = realmPath

	pull := NewPull(c.client, &cloneCfg)
	pull.SetSourceMode(true)
	_ = pull.ExecuteAll()

	fmt.Printf("\nRepository cloned successfully into '%s'!\n", repoName)
	return nil
}

func extractRepoName(realmPath string) string {
	realmPath = strings.TrimSuffix(realmPath, "/")

	parts := strings.Split(realmPath, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}
