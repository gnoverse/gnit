package main

import (
	"fmt"
	"strings"

	config "github.com/gnoverse/gnit"
	filesystem "github.com/gnoverse/gnit"
	gnokey "github.com/gnoverse/gnit"
)

type Pull struct {
	client      *gnokey.Client
	config      *config.Config
	sourceMode  bool
}

func NewPull(client *gnokey.Client, cfg *config.Config) *Pull {
	return &Pull{
		client:     client,
		config:     cfg,
		sourceMode: false,
	}
}

func (p *Pull) SetSourceMode(enabled bool) {
	p.sourceMode = enabled
}

func (p *Pull) Execute(filename string) error {
	if err := CheckGnitRepository(); err != nil {
		return err
	}

	fmt.Printf("Pulling '%s'...\n", filename)

	packageAlias := config.PackageAlias(p.config.RealmPath)
	query := fmt.Sprintf("%s.Repository.Pull(\"%s\")", packageAlias, filename)

	content, err := p.client.RunQuery(p.config.RealmPath, query)
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

	fmt.Println("Pulling all files from repository...")

	packageAlias := config.PackageAlias(p.config.RealmPath)
	query := fmt.Sprintf("%s.Repository.SerializePullAll()", packageAlias)

	serializedData, err := p.client.RunQuery(p.config.RealmPath, query)
	if err != nil {
		if p.sourceMode {
			fmt.Println("Repository not found or empty, trying to pull realm source files...")
			return p.pullRealmSource()
		}
		return fmt.Errorf("failed to pull files: %w", err)
	}

	if len(serializedData) == 0 {
		fmt.Println("No files found in repository")
		if p.sourceMode {
			return p.pullRealmSource()
		}
		return nil
	}

	files, err := parseSerializedFiles(string(serializedData))
	if err != nil {
		return fmt.Errorf("failed to parse files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No files found in repository")
		if p.sourceMode {
			return p.pullRealmSource()
		}
		return nil
	}

	fmt.Printf("Found %d file(s), writing to disk...\n", len(files))

	for filename, content := range files {
		if err := filesystem.WriteFile(filename, content); err != nil {
			return fmt.Errorf("failed to write '%s': %w", filename, err)
		}
		fmt.Printf("  pulled: %s (%d bytes)\n", filename, len(content))
	}

	fmt.Printf("\nSuccessfully pulled %d file(s)\n", len(files))

	if p.sourceMode {
		return p.pullRealmSource()
	}

	return nil
}

func (p *Pull) pullRealmSource() error {
	fmt.Println("\nFetching realm source files...")

	fileListOutput, err := p.client.QueryFileList(p.config.RealmPath)
	if err != nil {
		return fmt.Errorf("failed to list realm files: %w", err)
	}

	files := parseRealmFileList(fileListOutput)
	if len(files) == 0 {
		fmt.Println("No source files found in realm")
		return nil
	}

	fmt.Printf("Found %d source file(s), pulling...\n", len(files))

	for _, filename := range files {
		fullPath := p.config.RealmPath + "/" + filename
		content, err := p.client.QueryFileContent(fullPath)
		if err != nil {
			fmt.Printf("Warning: failed to pull '%s': %v\n", filename, err)
			continue
		}

		if err := filesystem.WriteFile(filename, []byte(content)); err != nil {
			return fmt.Errorf("failed to write '%s': %w", filename, err)
		}

		fmt.Printf("  pulled: %s (%d bytes)\n", filename, len(content))
	}

	return nil
}

func parseRealmFileList(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}
	}

	var files []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, ":") {
			continue
		}

		parts := strings.Split(line, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" && !strings.Contains(part, ":") {
				files = append(files, part)
			}
		}
	}

	return files
}

func parseSerializedFiles(data string) (map[string][]byte, error) {
	lines := strings.Split(data, "\n")
	files := make(map[string][]byte)

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}

		filename := unescapeString(parts[0])
		content := unescapeString(parts[1])
		files[filename] = []byte(content)
	}

	return files, nil
}

func unescapeString(s string) string {
	result := ""
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			next := s[i+1]
			if next == '\\' {
				result += "\\"
				i += 2
			} else if next == '|' {
				result += "|"
				i += 2
			} else if next == 'n' {
				result += "\n"
				i += 2
			} else {
				result += string(s[i])
				i++
			}
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
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
