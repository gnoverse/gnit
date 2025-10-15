package client

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	config *Config
}

func NewClient(cfg *Config) *Client {
	return &Client{config: cfg}
}

func (c *Client) QueryRaw(expression string) (string, error) {
	cmd := exec.Command("gnokey", "query", "vm/qeval",
		"-data", expression,
		"-remote", c.config.Remote)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	return extractDataLine(string(output))
}

func (c *Client) QueryFileList(realmPath string) (string, error) {
	cmd := exec.Command("gnokey", "query", "vm/qfile",
		"-data", realmPath,
		"-remote", c.config.Remote)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("query file list failed: %w", err)
	}

	return string(output), nil
}

func (c *Client) QueryFileContent(filePath string) (string, error) {
	cmd := exec.Command("gnokey", "query", "vm/qfile",
		"-data", filePath,
		"-remote", c.config.Remote)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("query file content failed: %w", err)
	}

	content := string(output)
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			fileContent := strings.TrimPrefix(line, "data: ")
			if i+1 < len(lines) {
				remainingLines := strings.Join(lines[i+1:], "\n")
				if remainingLines != "" {
					fileContent = fileContent + "\n" + remainingLines
				}
			}
			return fileContent, nil
		}
	}

	return content, nil
}

func (c *Client) Run(gnoCode string) error {
	tmpFile := "/tmp/gnit_tx.gno"
	if err := os.WriteFile(tmpFile, []byte(gnoCode), 0644); err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("gnokey", "maketx", "run",
		"-gas-fee", c.config.GasFee,
		"-gas-wanted", c.config.GasWanted,
		"-broadcast",
		"-chainid", c.config.ChainID,
		"-remote", c.config.Remote,
		c.config.Account,
		tmpFile)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

func (c *Client) RunQuery(realmPath string, expression string) ([]byte, error) {
	gnoCode := fmt.Sprintf(`package main

import (
	"%s"
)

func main() {
	result := %s
	content := make([]byte, len(result))
	copy(content, result)
	println(string(content))
}
`, realmPath, expression)

	tmpFile := "/tmp/gnit_query.gno"
	if err := os.WriteFile(tmpFile, []byte(gnoCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	cmd := exec.Command("gnokey", "maketx", "run",
		"-gas-fee", c.config.GasFee,
		"-gas-wanted", c.config.GasWanted,
		"-broadcast",
		"-chainid", c.config.ChainID,
		"-remote", c.config.Remote,
		c.config.Account,
		tmpFile)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run query failed: %w", err)
	}

	content := extractTransactionOutput(string(output))
	return []byte(content), nil
}

func extractTransactionOutput(output string) string {
	lines := strings.Split(output, "\n")
	var contentLines []string
	foundContent := false

	for _, line := range lines {
		if foundContent && isTransactionEndMetadata(line) {
			break
		}

		if foundContent {
			contentLines = append(contentLines, line)
			continue
		}

		if isTransactionMetadata(line) {
			continue
		}

		foundContent = true
		contentLines = append(contentLines, line)
	}

	return strings.Join(contentLines, "\n")
}

func isTransactionMetadata(line string) bool {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return true
	}

	metadataPrefixes := []string{
		"GAS WANTED:",
		"GAS USED:",
		"HEIGHT:",
		"EVENTS:",
		"INFO:",
		"TX HASH:",
	}

	for _, prefix := range metadataPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return strings.Contains(line, "OK!")
}

func isTransactionEndMetadata(line string) bool {
	metadataPrefixes := []string{
		"GAS WANTED:",
		"GAS USED:",
		"HEIGHT:",
		"EVENTS:",
		"INFO:",
		"TX HASH:",
	}

	for _, prefix := range metadataPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return strings.Contains(line, "OK!")
}

func extractDataLine(output string) (string, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			return line, nil
		}
	}
	return "", fmt.Errorf("data: line not found in output")
}
