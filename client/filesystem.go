package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CollectFiles() (map[string][]byte, error) {
	files := make(map[string][]byte)

	matcher, err := NewMatcher(".")
	if err != nil {
		return nil, fmt.Errorf("failed to load .gnitignore: %w", err)
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		cleanPath := filepath.Clean(path)
		if strings.HasPrefix(cleanPath, "./") {
			cleanPath = cleanPath[2:]
		}

		if matcher.Match(cleanPath) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Printf("Warning: unable to read %s: %v\n", path, readErr)
			return nil
		}

		files[cleanPath] = content
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

func WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

func SerializeFiles(files map[string][]byte) string {
	var builder strings.Builder
	for filename, content := range files {
		escaped := strings.ReplaceAll(string(content), "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		escaped = strings.ReplaceAll(escaped, "|", "\\|")

		builder.WriteString(filename)
		builder.WriteString("|")
		builder.WriteString(escaped)
		builder.WriteString("\n")
	}
	return builder.String()
}
