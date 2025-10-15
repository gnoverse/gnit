package main

import (
	"fmt"
	"os"
	"strings"

	"gnit/commands"
	"gnit/config"
	"gnit/gnokey"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.DefaultConfig()

	client := gnokey.NewClient(cfg)

	command := os.Args[1]

	switch command {
	case "clone":
		handleClone(client, cfg)
	case "add":
		handleAdd(cfg)
	case "pull":
		handlePull(client, cfg)
	case "commit":
		handleCommit(client, cfg)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printUsage()
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
	cmd := commands.NewClone(client, cfg)

	if err := cmd.Execute(realmPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleAdd(cfg *config.Config) {
	if len(os.Args) < 3 {
		fmt.Println("Error: files or directories required for add")
		fmt.Println("Usage: gnit add <file|directory>...")
		os.Exit(1)
	}

	paths := os.Args[2:]
	cmd := commands.NewAdd(cfg)

	if err := cmd.Execute(paths); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handlePull(client *gnokey.Client, cfg *config.Config) {
	cmd := commands.NewPull(client, cfg)

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
	if len(os.Args) < 3 {
		fmt.Println("Error: message required for commit")
		fmt.Println("Usage: gnit commit \"<message>\"")
		os.Exit(1)
	}

	message := strings.Join(os.Args[2:], " ")
	message = strings.Trim(message, "\"")

	cmd := commands.NewCommit(client, cfg)

	if err := cmd.Execute(message); err != nil {
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
	fmt.Println("  pull [options] [file]    Fetch file(s) from the repository")
	fmt.Println("    --source, -s           Also pull the realm source code to realm.gno")
	fmt.Println("  commit <message>         Commit staged changes with a message")
	fmt.Println("  help                     Display this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gnit clone gno.land/r/demo/myrepo")
	fmt.Println("  gnit add file.gno src/")
	fmt.Println("  gnit pull                    # Pull all files from repository")
	fmt.Println("  gnit pull example.gno        # Pull specific file")
	fmt.Println("  gnit pull -s                 # Pull all files + realm source code")
	fmt.Println("  gnit commit \"My commit message\"")
}
