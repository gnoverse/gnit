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

	listQuery := fmt.Sprintf("%s.Repository.ListFiles()", packageAlias)
	listData, err := s.client.RunQuery(s.config.RealmPath, listQuery)
	if err != nil {
		listData = []byte{}
	}

	committedFiles := make(map[string][]byte)

	if len(listData) > 0 {
		filenames, err := parseFileList(string(listData))
		if err == nil {
			for _, filename := range filenames {
				pullQuery := fmt.Sprintf("%s.Repository.Pull(\"%s\")", packageAlias, filename)
				content, err := s.client.RunQuery(s.config.RealmPath, pullQuery)
				if err != nil {
					continue
				}
				committedFiles[filename] = content
			}
		}
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

	var deleted []string
	for filename := range committedFiles {
		if _, exists := localFiles[filename]; !exists {
			if !stagedFiles[filename] {
				deleted = append(deleted, filename)
			}
		}
	}

	fmt.Printf("On branch %s\n", s.config.RealmPath)

	if len(staged) == 0 && len(modified) == 0 && len(untracked) == 0 && len(deleted) == 0 {
		fmt.Println("Your branch is up to date.")
		fmt.Println("nothing to commit, working tree clean")
		return nil
	}

	fmt.Println()

	if len(staged) > 0 {
		fmt.Println("Changes to be committed:")
		fmt.Println("  (use \"gnit restore --staged <file>...\" to unstage)")
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
		fmt.Println("  (use \"gnit restore <file>...\" to discard changes in working directory)")
		fmt.Println()
		for _, filename := range modified {
			fmt.Printf("\tmodified:   %s\n", filename)
		}
		fmt.Println()
	}

	if len(deleted) > 0 {
		fmt.Println("Deleted files:")
		fmt.Println("  (use \"gnit pull <file>...\" to restore)")
		fmt.Println()
		for _, filename := range deleted {
			fmt.Printf("\tdeleted:    %s\n", filename)
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

		if entry.Name() == ".gnit" {
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
