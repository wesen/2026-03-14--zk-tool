package obsidiancli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLineList(t *testing.T) {
	parsed, err := ParseOutput(OutputLineList, "\nalpha\nbeta\n\n")
	require.NoError(t, err)
	require.Equal(t, []string{"alpha", "beta"}, parsed)
}

func TestParseKeyValue(t *testing.T) {
	parsed, err := ParseOutput(OutputKeyValue, "name=one\npath=/tmp/example\n")
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"name": "one",
		"path": "/tmp/example",
	}, parsed)
}

func TestParseJSON(t *testing.T) {
	parsed, err := ParseOutput(OutputJSON, `{"ok":true,"count":3}`)
	require.NoError(t, err)
	result := parsed.(map[string]any)
	require.Equal(t, true, result["ok"])
	require.EqualValues(t, 3, result["count"])
}
