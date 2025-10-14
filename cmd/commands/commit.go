package commands

import (
	"fmt"
	"os"

	"gnit/config"
	"gnit/filesystem"
	"gnit/gnokey"
)

type Commit struct {
	client *gnokey.Client
	config *config.Config
}

func NewCommit(client *gnokey.Client, cfg *config.Config) *Commit {
	return &Commit{
		client: client,
		config: cfg,
	}
}

func (c *Commit) Execute(message string) error {
	if err := CheckGnitRepository(); err != nil {
		return err
	}

	fmt.Printf("Committing with message: '%s'...\n", message)

	gnitFile, err := ReadGnitFile()
	if err != nil {
		return fmt.Errorf("failed to read .gnit file: %w", err)
	}

	if len(gnitFile.StagedFiles) == 0 {
		return fmt.Errorf("no files staged for commit\nUse 'gnit add <file>' to stage files")
	}

	files := make(map[string][]byte)
	for _, filename := range gnitFile.StagedFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read '%s': %w", filename, err)
		}
		files[filename] = content
	}

	fmt.Printf("Files to commit: %d\n", len(files))
	for filename := range files {
		fmt.Printf("  - %s\n", filename)
	}

	filesData := filesystem.SerializeFiles(files)

	gnoCode := c.generateCommitCode(message, filesData)

	if err := c.client.Run(gnoCode); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	gnitFile.StagedFiles = []string{}
	if err := WriteGnitFileData(gnitFile); err != nil {
		fmt.Printf("Warning: failed to clear staged files: %v\n", err)
	}

	fmt.Printf("Commit successful!\n")
	return nil
}

func (c *Commit) generateCommitCode(message, filesData string) string {
	return fmt.Sprintf(`package main

import (
	"strings"
	"gno.land/r/example"
)

func unescape(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\|", "|")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

func main() {
	filesData := %q
	lines := strings.Split(filesData, "\n")
	files := make(map[string][]byte)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			files[parts[0]] = []byte(unescape(parts[1]))
		}
	}

	hash := example.Repo.Commit(%q, files)
	println("Commit hash:", hash)
}
`, filesData, message)
}
