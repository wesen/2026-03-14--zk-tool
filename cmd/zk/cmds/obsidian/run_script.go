package obsidian

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	zkcommon "github.com/wesen/2026-03-14--zk-tool/cmd/zk/cmds/common"
	"github.com/wesen/2026-03-14--zk-tool/pkg/obsidianjs"
)

// RunScriptCommand runs one JavaScript file with require("obsidian") preloaded.
type RunScriptCommand struct {
	*cmds.CommandDescription
}

// RunScriptSettings stores parsed command settings.
type RunScriptSettings struct {
	ScriptPath  string `glazed:"script"`
	BinaryPath  string `glazed:"binary"`
	Vault       string `glazed:"vault"`
	PrintResult bool   `glazed:"print-result"`
}

var _ cmds.BareCommand = (*RunScriptCommand)(nil)
var _ cmds.GlazeCommand = (*RunScriptCommand)(nil)

// NewRunScriptCommand creates the Glazed command.
func NewRunScriptCommand() (*RunScriptCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"run-script",
		cmds.WithShort("Run a JavaScript file against the local Obsidian CLI"),
		cmds.WithLong(`Execute a JavaScript file with require("obsidian") available.

The command configures the module to use ~/.local/bin/obsidian by default and
returns the script result as a structured row.

Examples:
  zk obsidian run-script scripts/js-tests/obsidian-version.js
  zk obsidian run-script scripts/js-tests/obsidian-sample-files.js --vault obsidian-vault
  zk obsidian run-script scripts/js-tests/obsidian-version.js --output json`),
		cmds.WithFlags(
			fields.New(
				"script",
				fields.TypeString,
				fields.WithIsArgument(true),
				fields.WithHelp("Path to the JavaScript file to execute"),
			),
			fields.New(
				"binary",
				fields.TypeString,
				fields.WithHelp("Path to the Obsidian CLI wrapper"),
			),
			fields.New(
				"vault",
				fields.TypeString,
				fields.WithHelp("Optional vault name override"),
			),
			fields.New(
				"print-result",
				fields.TypeBool,
				fields.WithDefault(true),
				fields.WithHelp("Print the script result in human mode"),
			),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)

	return &RunScriptCommand{CommandDescription: desc}, nil
}

func (c *RunScriptCommand) Run(ctx context.Context, vals *values.Values) error {
	settings := &RunScriptSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode command settings")
	}

	result, err := obsidianjs.RunFile(ctx, obsidianjs.RunOptions{
		ScriptPath: settings.ScriptPath,
		BinaryPath: settings.BinaryPath,
		Vault:      settings.Vault,
	})
	if err != nil {
		return err
	}

	if !settings.PrintResult {
		fmt.Printf("%s\n", result.ScriptPath)
		return nil
	}

	switch parsed := result.Parsed.(type) {
	case string:
		fmt.Println(parsed)
	default:
		raw, marshalErr := json.MarshalIndent(parsed, "", "  ")
		if marshalErr != nil {
			fmt.Println(result.Output)
		} else {
			fmt.Println(string(raw))
		}
	}

	return nil
}

func (c *RunScriptCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	settings := &RunScriptSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode command settings")
	}

	result, err := obsidianjs.RunFile(ctx, obsidianjs.RunOptions{
		ScriptPath: settings.ScriptPath,
		BinaryPath: settings.BinaryPath,
		Vault:      settings.Vault,
	})
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("script", result.ScriptPath),
		types.MRP("binary", result.BinaryPath),
		types.MRP("vault", result.Vault),
		types.MRP("result", result.Parsed),
		types.MRP("raw_result", result.Output),
		types.MRP("status", "ok"),
	)
	return gp.AddRow(ctx, row)
}

// NewRunScriptCobraCommand builds the Cobra adapter for the command.
func NewRunScriptCobraCommand() (*cobra.Command, error) {
	command, err := NewRunScriptCommand()
	if err != nil {
		return nil, fmt.Errorf("build run-script command: %w", err)
	}
	return zkcommon.BuildCobraCommand(command)
}
