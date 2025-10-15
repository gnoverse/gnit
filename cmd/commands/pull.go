package commands

import (
	"fmt"
	"strings"

	"gnit/config"
	"gnit/filesystem"
	"gnit/gnokey"
)

type Pull struct {
	client *gnokey.Client
	config *config.Config
}

func NewPull(client *gnokey.Client, cfg *config.Config) *Pull {
	return &Pull{
		client: client,
		config: cfg,
	}
}

func (p *Pull) Execute(filename string) error {
	if err := CheckGnitRepository(); err != nil {
		return err
	}

	fmt.Printf("Pulling '%s'...\n", filename)

	query := fmt.Sprintf("%s.Repo.Pull(\"%s\")", p.config.RealmPath, filename)
	content, err := p.client.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query file: %w", err)
	}

	if err := filesystem.WriteFile(filename, content); err != nil {
		return err
	}

	fmt.Printf("File '%s' fetched successfully (%d bytes)\n", filename, len(content))
	return nil
}

func (p *Pull) ExecuteAll() error {
	if err := CheckGnitRepository(); err != nil {
		return err
	}

	fmt.Println("Fetching list of files...")

	query := fmt.Sprintf("%s.Repo.ListFiles()", p.config.RealmPath)
	result, err := p.client.QueryRaw(query)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(result) == 0 {
		fmt.Println("No files found in repository")
		return nil
	}

	files, err := parseFileList(result)
	if err != nil {
		return fmt.Errorf("failed to parse file list: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No files found in repository")
		return nil
	}

	fmt.Printf("Found %d file(s), pulling all...\n", len(files))

	for _, filename := range files {
		if err := p.Execute(filename); err != nil {
			return fmt.Errorf("failed to pull '%s': %w", filename, err)
		}
	}

	fmt.Printf("\nSuccessfully pulled %d file(s)\n", len(files))
	return nil
}

func parseFileList(data string) ([]string, error) {
	str := strings.TrimSpace(data)

	str = strings.TrimPrefix(str, "data: ")

	sliceStart := strings.Index(str, "slice[")
	if sliceStart == -1 {
		return []string{}, nil
	}

	sliceStart += len("slice[")
	sliceEnd := strings.LastIndex(str, "]")
	if sliceEnd == -1 || sliceEnd <= sliceStart {
		return []string{}, nil
	}

	content := str[sliceStart:sliceEnd]
	if strings.TrimSpace(content) == "" {
		return []string{}, nil
	}

	var files []string
	var currentFile strings.Builder
	inQuotes := false
	escaped := false

	for i := 0; i < len(content); i++ {
		ch := content[i]

		if escaped {
			currentFile.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			if inQuotes {
				files = append(files, currentFile.String())
				currentFile.Reset()
				inQuotes = false
			} else {
				inQuotes = true
			}
			continue
		}

		if inQuotes {
			currentFile.WriteByte(ch)
		}
	}

	return files, nil
}
