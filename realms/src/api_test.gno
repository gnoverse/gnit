package gnit

import "testing"

func TestNewRepository(t *testing.T) {
	r := NewRepository("foo")

	if r.identity.Name != "foo" {
		t.Errorf("expected %s, got %s", "foo", r.identity.Name)
	}
}

func TestCommit(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"main.go":   []byte("package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"),
		"README.md": []byte("# Test Project\n\nThis is a test project."),
	}

	commitHash := r.Commit("Initial commit", files)

	if commitHash == "" {
		t.Error("expected non-empty commit hash")
	}

	commit := r.GetCommit(commitHash)
	if commit == nil {
		t.Error("expected commit to be stored")
	}

	if commit.Message != "Initial commit" {
		t.Errorf("expected message 'Initial commit', got %s", commit.Message)
	}

	if commit.Hash != commitHash {
		t.Errorf("expected hash %s, got %s", commitHash, commit.Hash)
	}
}

func TestGetFile(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"main.go":   []byte("package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"),
		"README.md": []byte("# Test Project\n\nThis is a test project."),
	}

	commitHash := r.Commit("Initial commit", files)

	mainContent := r.GetFile(commitHash, "main.go")
	if string(mainContent) != string(files["main.go"]) {
		t.Errorf("expected %s, got %s", string(files["main.go"]), string(mainContent))
	}

	readmeContent := r.GetFile(commitHash, "README.md")
	if string(readmeContent) != string(files["README.md"]) {
		t.Errorf("expected %s, got %s", string(files["README.md"]), string(readmeContent))
	}

	nonExistentContent := r.GetFile(commitHash, "nonexistent.txt")
	if nonExistentContent != nil {
		t.Error("expected nil for non-existent file")
	}
}

func TestGetCommitNonExistent(t *testing.T) {
	r := NewRepository("test-repo")

	commit := r.GetCommit("nonexistent-hash")
	if commit != nil {
		t.Error("expected nil for non-existent commit")
	}
}

func TestMultipleCommits(t *testing.T) {
	r := NewRepository("test-repo")

	files1 := map[string][]byte{
		"file1.txt": []byte("First version"),
	}
	hash1 := r.Commit("First commit", files1)

	files2 := map[string][]byte{
		"file1.txt": []byte("Second version"),
		"file2.txt": []byte("New file"),
	}
	hash2 := r.Commit("Second commit", files2)

	if hash1 == hash2 {
		t.Error("expected different hashes for different commits")
	}

	content1 := r.GetFile(hash1, "file1.txt")
	if string(content1) != "First version" {
		t.Errorf("expected 'First version', got %s", string(content1))
	}

	content2 := r.GetFile(hash2, "file1.txt")
	if string(content2) != "Second version" {
		t.Errorf("expected 'Second version', got %s", string(content2))
	}

	newFile := r.GetFile(hash2, "file2.txt")
	if string(newFile) != "New file" {
		t.Errorf("expected 'New file', got %s", string(newFile))
	}
}

func TestRenderHome(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"README.md":     []byte("# Test"),
		"src/api.gno":   []byte("package gnit"),
		"src/types.gno": []byte("package gnit"),
	}

	r.Commit("Initial commit", files)

	result := r.Render("")

	if result == "" {
		t.Error("expected non-empty result")
	}

	expectedSubstrings := []string{
		"# test-repo",
		"Branch:",
		"main",
		"Latest:",
		"Initial commit",
		"## Files",
		"📁",
		"src/",
		"📄",
		"README.md",
	}

	for i := 0; i < len(expectedSubstrings); i++ {
		if !contains(result, expectedSubstrings[i]) {
			t.Errorf("expected result to contain '%s', got: %s", expectedSubstrings[i], result)
		}
	}
}

func TestRenderDirectory(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"README.md":     []byte("# Test"),
		"src/api.gno":   []byte("package gnit"),
		"src/types.gno": []byte("package gnit"),
	}

	r.Commit("Initial commit", files)

	result := r.Render("src")

	if result == "" {
		t.Error("expected non-empty result")
	}

	expectedSubstrings := []string{
		"# test-repo/src",
		"## Files",
		"📄",
		"api.gno",
		"types.gno",
	}

	for i := 0; i < len(expectedSubstrings); i++ {
		if !contains(result, expectedSubstrings[i]) {
			t.Errorf("expected result to contain '%s', got: %s", expectedSubstrings[i], result)
		}
	}
}

func TestRenderFile(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"README.md":   []byte("# Test Project"),
		"src/api.gno": []byte("package gnit\n\nfunc NewRepo() {}"),
	}

	r.Commit("Initial commit", files)

	result := r.Render("README.md")

	if result == "" {
		t.Error("expected non-empty result")
	}

	expectedSubstrings := []string{
		"# README.md",
		"**Size:**",
		"```md",
		"# Test Project",
	}

	for i := 0; i < len(expectedSubstrings); i++ {
		if !contains(result, expectedSubstrings[i]) {
			t.Errorf("expected result to contain '%s', got: %s", expectedSubstrings[i], result)
		}
	}

	result2 := r.Render("src/api.gno")

	if !contains(result2, "```gno") {
		t.Errorf("expected gno code block")
	}

	if !contains(result2, "package gnit") {
		t.Errorf("expected file content")
	}
}

func TestRenderFileNotFound(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"README.md": []byte("# Test"),
	}

	r.Commit("Initial commit", files)

	result := r.Render("nonexistent.txt")

	if !contains(result, "File not found") {
		t.Errorf("expected 'File not found' message, got: %s", result)
	}
}

func TestRenderEmptyRepository(t *testing.T) {
	r := NewRepository("empty-repo")

	result := r.Render("")

	if !contains(result, "# empty-repo") {
		t.Errorf("expected repo name in header")
	}

	if !contains(result, "No commits yet") {
		t.Errorf("expected 'No commits yet' message")
	}
}

func TestListDirectory(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"README.md":          []byte("# Test"),
		"src/api.gno":        []byte("package gnit"),
		"src/types.gno":      []byte("package gnit"),
		"src/utils/hash.gno": []byte("package utils"),
	}

	r.Commit("Initial commit", files)

	filesResult, dirsResult := r.ListDirectory("")

	if len(filesResult) != 1 {
		t.Errorf("expected 1 file in root, got %d", len(filesResult))
	}

	if len(dirsResult) != 1 {
		t.Errorf("expected 1 dir in root, got %d", len(dirsResult))
	}

	if filesResult[0] != "README.md" {
		t.Errorf("expected README.md, got %s", filesResult[0])
	}

	if dirsResult[0] != "src" {
		t.Errorf("expected src, got %s", dirsResult[0])
	}

	filesResult2, dirsResult2 := r.ListDirectory("src")

	if len(filesResult2) != 2 {
		t.Errorf("expected 2 files in src, got %d", len(filesResult2))
	}

	if len(dirsResult2) != 1 {
		t.Errorf("expected 1 dir in src, got %d", len(dirsResult2))
	}
}

func TestIsDirectory(t *testing.T) {
	r := NewRepository("test-repo")

	files := map[string][]byte{
		"README.md":     []byte("# Test"),
		"src/api.gno":   []byte("package gnit"),
		"src/types.gno": []byte("package gnit"),
	}

	r.Commit("Initial commit", files)

	if !r.IsDirectory("") {
		t.Error("expected root to be a directory")
	}

	if !r.IsDirectory("src") {
		t.Error("expected src to be a directory")
	}

	if !r.IsDirectory("src/") {
		t.Error("expected src/ to be a directory")
	}

	if r.IsDirectory("README.md") {
		t.Error("expected README.md to not be a directory")
	}

	if r.IsDirectory("nonexistent") {
		t.Error("expected nonexistent to not be a directory")
	}
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
