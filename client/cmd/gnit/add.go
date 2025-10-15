package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config "github.com/gnoverse/gnit"
	ignore "github.com/gnoverse/gnit"
)

type Add struct {
	config *config.Config
}

func NewAdd(cfg *config.Config) *Add {
	return &Add{
		config: cfg,
	}
}

func (a *Add) Execute(paths []string) error {
	if err := CheckGnitRepository(); err != nil {
		return err
	}

	if len(paths) == 0 {
		return fmt.Errorf("no files or directories specified")
	}

	gnitFile, err := ReadGnitFile()
	if err != nil {
		return fmt.Errorf("failed to read .gnit file: %w", err)
	}

	matcher, err := ignore.NewMatcher(".")
	if err != nil {
		return fmt.Errorf("failed to load .gnitignore: %w", err)
	}

	filesToAdd := make(map[string]bool)
	for _, path := range paths {
		if err := a.collectFiles(path, matcher, filesToAdd); err != nil {
			return fmt.Errorf("failed to process '%s': %w", path, err)
		}
	}

	stagedMap := make(map[string]bool)
	for _, f := range gnitFile.StagedFiles {
		stagedMap[f] = true
	}

	addedCount := 0
	for file := range filesToAdd {
		if !stagedMap[file] {
			gnitFile.StagedFiles = append(gnitFile.StagedFiles, file)
			stagedMap[file] = true
			addedCount++
			fmt.Printf("  added: %s\n", file)
		}
	}

	if addedCount == 0 {
		fmt.Println("No new files to add")
		return nil
	}

	if err := WriteGnitFileData(gnitFile); err != nil {
		return fmt.Errorf("failed to update .gnit file: %w", err)
	}

	fmt.Printf("\n%d file(s) staged for commit\n", addedCount)
	return nil
}

func (a *Add) collectFiles(path string, matcher *ignore.Matcher, result map[string]bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "./") {
		cleanPath = cleanPath[2:]
	}

	if cleanPath == ".gnit" || matcher.Match(cleanPath) {
		return nil
	}

	if info.IsDir() {
		return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			cleanP := filepath.Clean(p)
			if strings.HasPrefix(cleanP, "./") {
				cleanP = cleanP[2:]
			}

			if cleanP != ".gnit" && !matcher.Match(cleanP) {
				result[cleanP] = true
			}

			return nil
		})
	} else {
		result[cleanPath] = true
	}

	return nil
}
