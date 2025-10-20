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
	parts := strings.Split(realmPath, "/")
	packageAlias := parts[len(parts)-1]

	queryExpression := strings.Replace(expression, packageAlias+".", realmPath+".", 1)

	if strings.Contains(expression, ".Repository.Pull(") {
		return c.QueryFileInChunks(queryExpression)
	}

	return c.QueryDirectInChunks(queryExpression)
}

func (c *Client) QueryFileInChunks(expression string) ([]byte, error) {
	sizeQuery := strings.Replace(expression, "Repository.Pull(", "Repository.GetFileSize(", 1)
	sizeOutput, err := c.QueryEval(sizeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	sizeOutput = strings.TrimPrefix(sizeOutput, "data: ")
	var size int
	if _, err := fmt.Sscanf(sizeOutput, "(%d int)", &size); err != nil {
		return nil, fmt.Errorf("failed to parse file size: %w", err)
	}

	if size < 0 {
		return nil, fmt.Errorf("file not found")
	}

	if size == 0 {
		return []byte{}, nil
	}

	var content []byte
	chunkSize := 200
	for offset := 0; offset < size; offset += chunkSize {
		chunkQuery := strings.Replace(expression, "Repository.Pull(", "Repository.GetFileChunk(", 1)
		chunkQuery = strings.Replace(chunkQuery, ")", fmt.Sprintf(", %d, %d)", offset, chunkSize), 1)

		chunkOutput, err := c.QueryEval(chunkQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to get chunk at offset %d: %w", offset, err)
		}

		chunk := extractStringFromQuery(chunkOutput)
		content = append(content, []byte(chunk)...)
	}

	return content, nil
}

func (c *Client) QueryEval(expression string) (string, error) {
	cmd := exec.Command("gnokey", "query", "vm/qeval",
		"-data", expression,
		"-remote", c.config.Remote)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	return extractDataLine(string(output))
}

func extractStringFromQuery(output string) string {
	output = strings.TrimPrefix(output, "data: ")

	startIdx := strings.Index(output, `("`)
	if startIdx == -1 {
		return ""
	}
	startIdx += 2

	endIdx := strings.LastIndex(output, `" string)`)
	if endIdx == -1 {
		return ""
	}

	content := output[startIdx:endIdx]

	var result strings.Builder
	for i := 0; i < len(content); i++ {
		if content[i] == '\\' && i+1 < len(content) {
			switch content[i+1] {
			case 'n':
				result.WriteByte('\n')
				i++
			case 't':
				result.WriteByte('\t')
				i++
			case 'r':
				result.WriteByte('\r')
				i++
			case '\\':
				result.WriteByte('\\')
				i++
			case '"':
				result.WriteByte('"')
				i++
			default:
				result.WriteByte(content[i])
			}
		} else {
			result.WriteByte(content[i])
		}
	}

	return result.String()
}

func (c *Client) QueryDirectInChunks(expression string) ([]byte, error) {
	output, err := c.QueryEval(expression)
	if err != nil {
		return nil, err
	}

	return []byte(output), nil
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
