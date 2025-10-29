package gnit

import (
	"strconv"
)

func createTreeHash(files map[string][]byte) string {
	var content string

	var keys []string
	for k := range files {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, filename := range keys {
		content += filename + string(files[filename])
	}

	return simpleHash(content)
}

func createObjectHash(content []byte) string {
	return simpleHash(string(content))
}

func createCommitHash(commit *Commit) string {
	content := commit.Tree + commit.Message + commit.Author.Name + commit.Author.Email + strconv.FormatInt(commit.Timestamp, 10)

	for _, parent := range commit.Parents {
		content += parent
	}

	return simpleHash(content)
}

func createTreeHashFromMap(tree map[string]string) string {
	var content string

	var keys []string
	for k := range tree {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, filename := range keys {
		content += filename + tree[filename]
	}

	return simpleHash(content)
}

func simpleHash(input string) string {
	var hash int64 = 5381

	for i := 0; i < len(input); i++ {
		hash = ((hash << 5) + hash) + int64(input[i])
	}

	if hash < 0 {
		hash = -hash
	}

	return strconv.FormatInt(hash, 16)
}

func hasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

func hasSuffix(s, suffix string) bool {
	if len(suffix) > len(s) {
		return false
	}
	offset := len(s) - len(suffix)
	for i := 0; i < len(suffix); i++ {
		if s[offset+i] != suffix[i] {
			return false
		}
	}
	return true
}

func trimSuffix(s, suffix string) string {
	if hasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	
	var parts []string
	current := ""
	
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(path[i])
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "/" + parts[i]
	}
	return result
}

func (r *Repository) ListDirectory(dirPath string) (files []string, dirs []string) {
	dirPath = trimSuffix(dirPath, "/")
	if dirPath != "" {
		dirPath += "/"
	}
	
	allFiles := r.ListFiles()
	seenDirs := make(map[string]bool)
	
	for i := 0; i < len(allFiles); i++ {
		file := allFiles[i]
		
		if dirPath == "" {
			parts := splitPath(file)
			if len(parts) == 1 {
				files = append(files, file)
			} else if len(parts) > 1 {
				if !seenDirs[parts[0]] {
					dirs = append(dirs, parts[0])
					seenDirs[parts[0]] = true
				}
			}
		} else {
			if hasPrefix(file, dirPath) {
				relative := file[len(dirPath):]
				parts := splitPath(relative)
				
				if len(parts) == 1 {
					files = append(files, relative)
				} else if len(parts) > 1 {
					if !seenDirs[parts[0]] {
						dirs = append(dirs, parts[0])
						seenDirs[parts[0]] = true
					}
				}
			}
		}
	}
	
	return files, dirs
}

func (r *Repository) IsDirectory(path string) bool {
	path = trimSuffix(path, "/")
	if path == "" {
		return true
	}
	
	path += "/"
	allFiles := r.ListFiles()
	
	for i := 0; i < len(allFiles); i++ {
		if hasPrefix(allFiles[i], path) {
			return true
		}
	}
	
	return false
}

func formatBytes(size int) string {
	if size < 1024 {
		return strconv.Itoa(size) + " B"
	}
	if size < 1024*1024 {
		return strconv.Itoa(size/1024) + " KB"
	}
	return strconv.Itoa(size/(1024*1024)) + " MB"
}
