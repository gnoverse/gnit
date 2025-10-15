package main

import (
	"fmt"

	config "github.com/gnoverse/gnit"
	filesystem "github.com/gnoverse/gnit"
	gnokey "github.com/gnoverse/gnit"
)

type Restore struct {
	client *gnokey.Client
	config *config.Config
	staged bool
}

func NewRestore(client *gnokey.Client, cfg *config.Config) *Restore {
	return &Restore{
		client: client,
		config: cfg,
		staged: false,
	}
}

func (r *Restore) SetStaged(staged bool) {
	r.staged = staged
}

func (r *Restore) Execute(paths []string) error {
	if err := CheckGnitRepository(); err != nil {
		return err
	}

	if r.staged {
		return r.restoreStaged(paths)
	}

	return r.restoreWorkingTree(paths)
}

func (r *Restore) restoreStaged(paths []string) error {
	gnitFile, err := ReadGnitFile()
	if err != nil {
		return fmt.Errorf("failed to read .gnit file: %w", err)
	}

	if len(paths) == 0 {
		if len(gnitFile.StagedFiles) == 0 {
			fmt.Println("No files are staged")
			return nil
		}

		count := len(gnitFile.StagedFiles)
		gnitFile.StagedFiles = []string{}

		if err := WriteGnitFileData(gnitFile); err != nil {
			return fmt.Errorf("failed to update .gnit file: %w", err)
		}

		fmt.Printf("Unstaged %d file(s)\n", count)
		return nil
	}

	stagedMap := make(map[string]bool)
	for _, file := range gnitFile.StagedFiles {
		stagedMap[file] = true
	}

	unstagedCount := 0
	var newStagedFiles []string

	for _, file := range gnitFile.StagedFiles {
		shouldUnstage := false
		for _, path := range paths {
			if file == path {
				shouldUnstage = true
				break
			}
		}

		if shouldUnstage {
			unstagedCount++
			fmt.Printf("  unstaged: %s\n", file)
		} else {
			newStagedFiles = append(newStagedFiles, file)
		}
	}

	if unstagedCount == 0 {
		fmt.Println("No matching staged files found")
		return nil
	}

	gnitFile.StagedFiles = newStagedFiles
	if err := WriteGnitFileData(gnitFile); err != nil {
		return fmt.Errorf("failed to update .gnit file: %w", err)
	}

	fmt.Printf("\n%d file(s) unstaged\n", unstagedCount)
	return nil
}

func (r *Restore) restoreWorkingTree(paths []string) error {
	packageAlias := config.PackageAlias(r.config.RealmPath)
	query := fmt.Sprintf("%s.Repository.SerializePullAll()", packageAlias)

	serializedData, err := r.client.RunQuery(r.config.RealmPath, query)
	if err != nil {
		return fmt.Errorf("failed to fetch repository files: %w", err)
	}

	committedFiles := make(map[string][]byte)
	if len(serializedData) > 0 {
		committedFiles, err = parseSerializedFiles(string(serializedData))
		if err != nil {
			return fmt.Errorf("failed to parse repository files: %w", err)
		}
	}

	if len(paths) == 0 {
		if len(committedFiles) == 0 {
			fmt.Println("No files found in repository to restore")
			return nil
		}

		fmt.Printf("Restoring %d file(s) from repository...\n", len(committedFiles))

		for filename, content := range committedFiles {
			if err := filesystem.WriteFile(filename, content); err != nil {
				return fmt.Errorf("failed to restore '%s': %w", filename, err)
			}
			fmt.Printf("  restored: %s (%d bytes)\n", filename, len(content))
		}

		fmt.Printf("\nSuccessfully restored %d file(s)\n", len(committedFiles))
		return nil
	}

	restoredCount := 0
	for _, path := range paths {
		content, exists := committedFiles[path]
		if !exists {
			fmt.Printf("Warning: '%s' not found in repository\n", path)
			continue
		}

		if err := filesystem.WriteFile(path, content); err != nil {
			return fmt.Errorf("failed to restore '%s': %w", path, err)
		}

		fmt.Printf("  restored: %s (%d bytes)\n", path, len(content))
		restoredCount++
	}

	if restoredCount == 0 {
		fmt.Println("No files were restored")
		return nil
	}

	fmt.Printf("\nSuccessfully restored %d file(s)\n", restoredCount)
	return nil
}
