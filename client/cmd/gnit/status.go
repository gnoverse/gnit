package main

import (
	"fmt"
	"os"

	config "github.com/gnoverse/gnit"
	gnokey "github.com/gnoverse/gnit"
)

type Status struct {
	client *gnokey.Client
	config *config.Config
}

func NewStatus(client *gnokey.Client, cfg *config.Config) *Status {
	return &Status{
		client: client,
		config: cfg,
	}
}

func (s *Status) Execute() error {
	if err := CheckGnitRepository(); err != nil {
		return err
	}

	gnitFile, err := ReadGnitFile()
	if err != nil {
		return fmt.Errorf("failed to read .gnit file: %w", err)
	}

	packageAlias := config.PackageAlias(s.config.RealmPath)
	query := fmt.Sprintf("%s.Repository.SerializePullAll()", packageAlias)

	serializedData, err := s.client.RunQuery(s.config.RealmPath, query)
	if err != nil {
		serializedData = []byte{}
	}

	committedFiles := make(map[string][]byte)
	if len(serializedData) > 0 {
		committedFiles, _ = parseSerializedFiles(string(serializedData))
	}

	localFiles, err := getLocalFiles()
	if err != nil {
		return fmt.Errorf("failed to get local files: %w", err)
	}

	stagedFiles := make(map[string]bool)
	for _, filename := range gnitFile.StagedFiles {
		stagedFiles[filename] = true
	}

	var staged []string
	var modified []string
	var untracked []string

	for filename := range stagedFiles {
		staged = append(staged, filename)
	}

	for filename, localContent := range localFiles {
		committedContent, existsInCommit := committedFiles[filename]
		_, isStaged := stagedFiles[filename]

		if !existsInCommit && !isStaged {
			untracked = append(untracked, filename)
		} else if existsInCommit && !isStaged {
			if !bytesEqual(localContent, committedContent) {
				modified = append(modified, filename)
			}
		}
	}

	fmt.Printf("On branch %s\n", s.config.RealmPath)
	fmt.Println()

	if len(staged) > 0 {
		fmt.Println("Changes to be committed:")
		fmt.Println("  (use \"gnit reset <file>...\" to unstage)")
		fmt.Println()
		for _, filename := range staged {
			_, existsInCommit := committedFiles[filename]
			if existsInCommit {
				fmt.Printf("\tmodified:   %s\n", filename)
			} else {
				fmt.Printf("\tnew file:   %s\n", filename)
			}
		}
		fmt.Println()
	}

	if len(modified) > 0 {
		fmt.Println("Changes not staged for commit:")
		fmt.Println("  (use \"gnit add <file>...\" to update what will be committed)")
		fmt.Println()
		for _, filename := range modified {
			fmt.Printf("\tmodified:   %s\n", filename)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		fmt.Println("  (use \"gnit add <file>...\" to include in what will be committed)")
		fmt.Println()
		for _, filename := range untracked {
			fmt.Printf("\t%s\n", filename)
		}
		fmt.Println()
	}

	if len(staged) == 0 && len(modified) == 0 && len(untracked) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	}

	return nil
}

func getLocalFiles() (map[string][]byte, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	files := make(map[string][]byte)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if entry.Name() == ".gnit" || entry.Name() == "gnomod.toml" {
			continue
		}

		content, err := os.ReadFile(entry.Name())
		if err != nil {
			continue
		}

		files[entry.Name()] = content
	}

	return files, nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
