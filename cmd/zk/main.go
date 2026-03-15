package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	obscmd "github.com/wesen/2026-03-14--zk-tool/cmd/zk/cmds/obsidian"
	"github.com/spf13/cobra"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	rootCmd := &cobra.Command{
		Use:   "zk",
		Short: "ZK tool command-line utilities",
	}

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
