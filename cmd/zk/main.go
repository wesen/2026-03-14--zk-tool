package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	obscmd "github.com/wesen/2026-03-14--zk-tool/cmd/zk/cmds/obsidian"
	"github.com/wesen/2026-03-14--zk-tool/pkg/doc"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	rootCmd := &cobra.Command{
		Use:   "zk",
		Short: "ZK tool command-line utilities",
	}

	helpSystem := help.NewHelpSystem()
	if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load help docs: %v\n", err)
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	obsidianCmd, err := obscmd.NewCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building obsidian command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(obsidianCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
