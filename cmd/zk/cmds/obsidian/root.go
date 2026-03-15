package obsidian

import "github.com/spf13/cobra"

// NewCommand creates the obsidian command group.
func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "obsidian",
		Short: "Run Obsidian-backed commands for the ZK tool",
	}

	runScriptCmd, err := NewRunScriptCobraCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(runScriptCmd)

	return root, nil
}
