package main

import (
	"fmt"

	config "github.com/gnoverse/gnit"
	gnokey "github.com/gnoverse/gnit"
)

type Init struct {
	client *gnokey.Client
	config *config.Config
}

func NewInit(client *gnokey.Client, cfg *config.Config) *Init {
	return &Init{
		client:     client,
		config:     cfg,
	}
}

func (i *Init) Execute() error {
	fmt.Printf("Init '%s' repository...\n", i.config.RealmPath)

	gnitFile := config.GnitFile{
		StagedFiles: []string{},
	}

	if err := WriteGnitFileData(&gnitFile); err != nil {
		return fmt.Errorf("failed to create .gnit file: %w", err)
	}

	fmt.Printf("Initialized gnit repository with realm: %s\n", i.config.RealmPath)
	return nil
}
