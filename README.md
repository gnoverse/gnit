# gnit

Git-like version control for Gno realms. Store and manage code directly on-chain.

## Getting Started

### Installation

```bash
make install
```

### Quick Start

```bash
# Clone a realm repository
gnit clone gno.land/r/demo/myrepo

# Make changes to files, then stage them
gnit add file.gno src/

# Commit to the realm
gnit commit "Add new feature"

# Pull files from repository
gnit pull
```

### Setup a New Repository

In your realm:

```go
package myrepo

import "gno.land/p/demo/gnit"

var Repository *gnit.Repository

func init() {
    Repository = gnit.NewRepository("myrepo")
}
```

Now you can commit files to your realm using `gnit`.

## How It Works

**Two parts:**
- **Gno package**: On-chain repository using AVL trees for storage
- **CLI tool**: Interacts with realms via `gnokey`

**Object model:**
```
Commit → Tree → Blob
```

Each commit points to a tree (file path → blob hash map). Trees point to blobs (file content).

## CLI Reference

```bash
gnit clone <realm-path>          # Clone repository (creates directory)
gnit add <files>...              # Stage files for commit
gnit commit "<message>"          # Commit staged files to realm
gnit pull                        # Pull all files from HEAD
gnit pull <file>                 # Pull specific file
gnit pull --source               # Pull files + realm source code
```

## Configuration

Default config (for local testing):
```
Remote:     tcp://127.0.0.1:26657
ChainID:    dev
Account:    test
RealmPath:  (from gnomod.toml)
```

## Current Status

**Implemented:**
- ✅ Commit files to realm
- ✅ Pull files from repository
- ✅ Clone repositories
- ✅ Stage files with `.gnitignore` support
- ✅ Content-addressed storage (DJB2 hash)
- ✅ Single branch (main)

**Not Yet:**
- ❌ Multiple branches
- ❌ Commit history/log
- ❌ Merge operations
- ❌ Nested directories (flat tree structure)
- ❌ SHA-1 hashing (uses DJB2)
- ❌ Parent commit tracking

## API

### Gno Package

```go
// Create repository
func NewRepository(name string) *Repository

// Write operations
func (r *Repository) Commit(message string, files map[string][]byte) string

// Read operations
func (r *Repository) Pull(file string) []byte
func (r *Repository) GetCommit(hash string) *Commit
func (r *Repository) GetFile(commitHash, path string) []byte
func (r *Repository) GetHeadCommit() *Commit
func (r *Repository) ListFiles() []string
func (r *Repository) GetCurrentBranch() string
```

### Types

```go
type Commit struct {
    Hash      string
    Tree      string
    Parents   []string
    Author    Identity
    Committer Identity
    Message   string
    Timestamp int64
}

type Identity struct {
    Name    string
    Email   string
    Address address
}
```

## Roadmap

### Next Steps
- **Render function** for on-chain web UI
- **Improved tree handling** (hierarchical directories)
- **Repository explorer** (browse commits, files, diffs)
