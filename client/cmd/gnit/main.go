package main

import (
	"fmt"
	"os"
	"strings"

	config "github.com/gnoverse/gnit"
	gnokey "github.com/gnoverse/gnit"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.DefaultConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	client := gnokey.NewClient(cfg)

	command := os.Args[1]

	switch command {
	case "init":
		handleInit(client, cfg)
	case "clone":
		handleClone(client, cfg)
	case "add":
		handleAdd(cfg)
	case "status":
		handleStatus(client, cfg)
	case "pull":
		handlePull(client, cfg)
	case "commit":
		handleCommit(client, cfg)
	case "restore":
		handleRestore(client, cfg)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleInit(client *gnokey.Client, cfg *config.Config) {
	cmd := NewInit(client, cfg)

	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleClone(client *gnokey.Client, cfg *config.Config) {
	if len(os.Args) < 3 {
		fmt.Println("Error: realm path required for clone")
		fmt.Println("Usage: gnit clone <realm-path>")
		os.Exit(1)
	}

	realmPath := os.Args[2]
	cmd := NewClone(client, cfg)

	if err := cmd.Execute(realmPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleAdd(cfg *config.Config) {
	if err := cfg.ValidateRealmPath(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("Error: files or directories required for add")
		fmt.Println("Usage: gnit add <file|directory>...")
		os.Exit(1)
	}

	paths := os.Args[2:]
	cmd := NewAdd(cfg)

	if err := cmd.Execute(paths); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleStatus(client *gnokey.Client, cfg *config.Config) {
	if err := cfg.ValidateRealmPath(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	cmd := NewStatus(client, cfg)

	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handlePull(client *gnokey.Client, cfg *config.Config) {
	if err := cfg.ValidateRealmPath(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	cmd := NewPull(client, cfg)

	sourceMode := false
	args := []string{}
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--source" || arg == "-s" {
			sourceMode = true
		} else {
			args = append(args, arg)
		}
	}

	if sourceMode {
		cmd.SetSourceMode(true)
	}

	if len(args) == 0 {
		if err := cmd.ExecuteAll(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	filename := args[0]
	if err := cmd.Execute(filename); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleCommit(client *gnokey.Client, cfg *config.Config) {
	if err := cfg.ValidateRealmPath(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("Error: message required for commit")
		fmt.Println("Usage: gnit commit \"<message>\"")
		os.Exit(1)
	}

	message := strings.Join(os.Args[2:], " ")
	message = strings.Trim(message, "\"")

	cmd := NewCommit(client, cfg)

	if err := cmd.Execute(message); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleRestore(client *gnokey.Client, cfg *config.Config) {
	if err := cfg.ValidateRealmPath(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	cmd := NewRestore(client, cfg)

	staged := false
	var paths []string

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--staged" || arg == "-s" {
			staged = true
		} else {
			paths = append(paths, arg)
		}
	}

	if staged {
		cmd.SetStaged(true)
	}

	if err := cmd.Execute(paths); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: gnit <command> [options]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  clone <realm-path>       Clone a repository from a realm path")
	fmt.Println("  add <file|directory>...  Stage files or directories for commit")
	fmt.Println("  status                   Show the working tree status")
	fmt.Println("  pull [options] [file]    Fetch file(s) from the repository")
	fmt.Println("    --source, -s           Also pull the realm source code to realm.gno")
	fmt.Println("  commit <message>         Commit staged changes with a message")
	fmt.Println("  restore [options] [file] Restore working tree files or unstage files")
	fmt.Println("    --staged, -s           Restore files in the staging area (unstage)")
	fmt.Println("  help                     Display this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gnit clone gno.land/r/demo/myrepo")
	fmt.Println("  gnit add file.gno src/")
	fmt.Println("  gnit status                  # Show working tree status")
	fmt.Println("  gnit pull                    # Pull all files from repository")
	fmt.Println("  gnit pull example.gno        # Pull specific file")
	fmt.Println("  gnit pull -s                 # Pull all files + realm source code")
	fmt.Println("  gnit commit \"My commit message\"")
	fmt.Println("  gnit restore file.gno        # Restore file from repository")
	fmt.Println("  gnit restore --staged file.gno # Unstage file")
	fmt.Println("  gnit restore --staged        # Unstage all files")
}
