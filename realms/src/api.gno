package gnit

import (
	"gno.land/p/nt/avl"
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
	if path == "" {
		files := r.ListFiles()
		result := "# Repository Files\n\n"
		for i := 0; i < len(files); i++ {
			result += "- " + files[i] + "\n"
		}
		return result
	}

	content := r.Pull(path)
	if content == nil {
		return "File not found: " + path
	}
	return string(content)
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
