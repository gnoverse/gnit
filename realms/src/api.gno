package gnit

import (
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

func NewRepository(name string) *Repository {
	return &Repository{
		identity: Identity{
			Name: name,
		},
		head: "main",
	}
}

func (r *Repository) Pull(file string) []byte {
	headCommit := r.GetHeadCommit()
	if headCommit == nil {
		return nil
	}

	return r.GetFile(headCommit.Hash, file)
}

func (r *Repository) Commit(message string, files map[string][]byte) string {
	if r.commits == nil {
		r.commits = avl.NewTree()
	}
	if r.objects == nil {
		r.objects = avl.NewTree()
	}
	if r.refs == nil {
		r.refs = avl.NewTree()
	}

	tree := make(map[string]string)

	headCommit := r.GetHeadCommit()
	if headCommit != nil {
		prevTreeValue, exists := r.objects.Get(headCommit.Tree)
		if exists {
			prevTree := prevTreeValue.(map[string]string)
			for path, objectHash := range prevTree {
				tree[path] = objectHash
			}
		}
	}

	for path, content := range files {
		objectHash := createObjectHash(content)
		r.objects.Set(objectHash, content)
		tree[path] = objectHash
	}

	treeHash := createTreeHashFromMap(tree)
	r.objects.Set(treeHash, tree)

	timestamp := int64(12345)

	parents := []string{}
	if headCommit != nil {
		parents = append(parents, headCommit.Hash)
	}

	commit := &Commit{
		Tree:      treeHash,
		Parents:   parents,
		Author:    r.identity,
		Committer: r.identity,
		Message:   message,
		Timestamp: timestamp,
	}

	commitHash := createCommitHash(commit)
	commit.Hash = commitHash

	r.commits.Set(commitHash, commit)
	r.refs.Set(r.head, commitHash)

	return commitHash
}

func (r *Repository) GetCommit(hash string) *Commit {
	if r.commits == nil {
		return nil
	}
	value, exists := r.commits.Get(hash)
	if !exists {
		return nil
	}
	return value.(*Commit)
}

func (r *Repository) GetFile(commitHash, path string) []byte {
	commit := r.GetCommit(commitHash)
	if commit == nil {
		return nil
	}

	treeValue, exists := r.objects.Get(commit.Tree)
	if !exists {
		return nil
	}

	tree := treeValue.(map[string]string)

	objectHash, exists := tree[path]
	if !exists {
		return nil
	}

	fileValue, exists := r.objects.Get(objectHash)
	if !exists {
		return nil
	}

	return fileValue.([]byte)
}

func (r *Repository) GetHeadCommit() *Commit {
	if r.refs == nil {
		return nil
	}

	commitHash, exists := r.refs.Get(r.head)
	if !exists {
		return nil
	}

	return r.GetCommit(commitHash.(string))
}

func (r *Repository) GetCurrentBranch() string {
	return r.head
}

func (r *Repository) ListFiles() []string {
	headCommit := r.GetHeadCommit()
	if headCommit == nil {
		return []string{}
	}

	treeValue, exists := r.objects.Get(headCommit.Tree)
	if !exists {
		return []string{}
	}

	tree := treeValue.(map[string]string)
	files := make([]string, 0, len(tree))

	for path := range tree {
		files = append(files, path)
	}

	return files
}

func (r *Repository) PullAll() map[string][]byte {
	headCommit := r.GetHeadCommit()
	if headCommit == nil {
		return map[string][]byte{}
	}

	treeValue, exists := r.objects.Get(headCommit.Tree)
	if !exists {
		return map[string][]byte{}
	}

	tree := treeValue.(map[string]string)
	result := make(map[string][]byte)

	for path, objectHash := range tree {
		fileValue, exists := r.objects.Get(objectHash)
		if exists {
			result[path] = fileValue.([]byte)
		}
	}

	return result
}

func (r *Repository) SerializePullAll() []byte {
	files := r.PullAll()

	var result string
	for path, content := range files {
		escapedPath := escapeString(path)
		escapedContent := escapeString(string(content))
		result += escapedPath + "|" + escapedContent + "\n"
	}

	return []byte(result)
}

func escapeString(s string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' {
			result += "\\\\"
		} else if c == '|' {
			result += "\\|"
		} else if c == '\n' {
			result += "\\n"
		} else {
			result += string(c)
		}
	}
	return result
}

func (r *Repository) Render(path string) string {
	path = trimSuffix(path, "/")

	if path == "" {
		return r.renderHome()
	}

	if r.IsDirectory(path) {
		return r.renderDirectory(path)
	}

	return r.renderFile(path)
}

func (r *Repository) renderHome() string {
	addr := strings.TrimSuffix(runtime.CurrentRealm().PkgPath(), "/")
	addr = addr[strings.Index(addr, "/r/"):]
	result := "# " + r.identity.Name + "\n\n"

	headCommit := r.GetHeadCommit()
	if headCommit != nil {
		result += "**Branch:** " + r.head + " | "
		result += "**Latest:** " + headCommit.Hash[:8] + " - \"" + headCommit.Message + "\"\n\n"
	} else {
		result += "**Branch:** " + r.head + " | No commits yet\n\n"
	}

	files, dirs := r.ListDirectory("")

	totalItems := len(files) + len(dirs)
	if totalItems == 0 {
		result += "_Repository is empty_\n"
		return result
	}

	result += "## Files (" + strconv.Itoa(totalItems) + ")\n\n"

	for i := 0; i < len(dirs); i++ {
		result += ufmt.Sprintf("[%s %s/](%s:%s/)\n\n", "📁", dirs[i], addr, dirs[i])
	}

	for i := 0; i < len(files); i++ {
		size := r.GetFileSize(files[i])
		result += ufmt.Sprintf("[%s %s](%s:%s)", "📄", files[i], addr, files[i])
		result += " - " + formatBytes(size)
		result += "\n\n"
	}

	return result
}

func (r *Repository) renderDirectory(path string) string {
	addr := strings.TrimSuffix(runtime.CurrentRealm().PkgPath(), "/")
	addr = addr[strings.Index(addr, "/r/"):]

	displayPath := path
	if displayPath == "" {
		displayPath = "/"
	} else {
		displayPath = "/" + displayPath
	}

	result := "# " + r.identity.Name + displayPath + "\n\n"
	result += "[← Back](" + addr + ")\n\n"

	files, dirs := r.ListDirectory(path)

	totalItems := len(files) + len(dirs)
	if totalItems == 0 {
		result += "_Empty directory_\n"
		return result
	}

	result += "## Files (" + strconv.Itoa(totalItems) + ")\n\n"

	for i := 0; i < len(dirs); i++ {
		result += ufmt.Sprintf("[%s %s/](%s%s:%s/)\n\n", "📁", dirs[i], addr, path, dirs[i])
	}

	for i := 0; i < len(files); i++ {
		size := r.GetFileSize(files[i])
		result += ufmt.Sprintf("[%s %s](%s:%s/%s)", "📄", files[i], addr, path, files[i])
		result += " - " + formatBytes(size)
		result += "\n\n"
	}

	return result
}

func (r *Repository) renderFile(path string) string {
	addr := strings.TrimSuffix(runtime.CurrentRealm().PkgPath(), "/")
	addr = addr[strings.Index(addr, "/r/"):]

	content := r.Pull(path)
	if content == nil {
		return "# File not found\n\nThe file `" + path + "` does not exist in this repository."
	}

	result := "# " + path + "\n\n"

	parts := splitPath(path)
	if len(parts) > 1 {
		parentPath := ""
		for i := 0; i < len(parts)-1; i++ {
			if i > 0 {
				parentPath += "/"
			}
			parentPath += parts[i]
		}
		result += "[← Back to " + parentPath + "](" + addr + ":" + parentPath + ")\n\n"
	} else {
		result += "[← Back](" + addr + ")\n\n"
	}

	size := len(content)
	result += "**Size:** " + formatBytes(size) + "\n\n"

	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			ext = path[i+1:]
			break
		}
		if path[i] == '/' {
			break
		}
	}

	if ext == "gno" || ext == "go" || ext == "md" || ext == "json" || ext == "yaml" || ext == "toml" {
		result += "```" + ext + "\n"
		result += string(content)
		if !hasSuffix(string(content), "\n") {
			result += "\n"
		}
		result += "```\n"
	} else {
		result += "```\n"
		result += string(content)
		if !hasSuffix(string(content), "\n") {
			result += "\n"
		}
		result += "```\n"
	}

	return result
}

func (r *Repository) GetFileChunk(filename string, offset, size int) string {
	content := r.Pull(filename)
	if content == nil {
		return ""
	}

	end := offset + size
	if end > len(content) {
		end = len(content)
	}
	if offset >= len(content) {
		return ""
	}

	return string(content[offset:end])
}

func (r *Repository) GetFileSize(filename string) int {
	content := r.Pull(filename)
	if content == nil {
		return -1
	}
	return len(content)
}
