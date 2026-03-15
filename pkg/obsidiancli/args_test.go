package obsidiancli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildArgsIncludesVaultFlagsAndPositionals(t *testing.T) {
	args, err := BuildArgs(Config{Vault: "Work"}, SpecSearch, CallOptions{
		Parameters: map[string]any{"query": "obsidian", "limit": 5},
		Flags:      []string{"json", "case", "json"},
		Positional: []string{"tail"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{
		"vault=Work",
		SpecSearch.Name,
		"limit=5",
		"query=obsidian",
		"case",
		"json",
		"tail",
	}, args)
}

func TestBuildArgsSkipsEmptyValues(t *testing.T) {
	args, err := BuildArgs(Config{}, SpecFilesList, CallOptions{
		Parameters: map[string]any{"folder": "", "ext": nil},
	})
	require.NoError(t, err)
	require.Equal(t, []string{SpecFilesList.Name}, args)
}
