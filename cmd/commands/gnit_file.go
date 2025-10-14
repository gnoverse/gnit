package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gnit/config"
)

func CheckGnitRepository() error {
	if _, err := os.Stat(".gnit"); os.IsNotExist(err) {
		return fmt.Errorf("not a gnit repository (no .gnit file found)\nRun 'gnit clone <realm-path>' to initialize a repository")
	}
	return nil
}

func ReadGnitFile() (*config.GnitFile, error) {
	data, err := os.ReadFile(".gnit")
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(data))

	var gnitFile config.GnitFile
	if err := json.Unmarshal([]byte(content), &gnitFile); err == nil {
		return &gnitFile, nil
	}

	if content != "" {
		return &config.GnitFile{
			RealmPath:   content,
			StagedFiles: []string{},
		}, nil
	}

	return nil, fmt.Errorf("invalid .gnit file format")
}

func WriteGnitFileData(gf *config.GnitFile) error {
	data, err := json.MarshalIndent(gf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(".gnit", data, 0644)
}
