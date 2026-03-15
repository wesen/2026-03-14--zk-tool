package obsidiancli

import (
	"context"
	"errors"
	osexec "os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeExecutor struct {
	invocation Invocation
	result     ExecResult
	err        error
}

func (f *fakeExecutor) Run(_ context.Context, inv Invocation) (ExecResult, error) {
	f.invocation = inv
	return f.result, f.err
}

func TestRunnerInvokesExecutor(t *testing.T) {
	executor := &fakeExecutor{
		result: ExecResult{Stdout: "alpha\nbeta\n"},
	}
	runner := NewRunner(Config{BinaryPath: "/tmp/obsidian", Vault: "Work"}, executor)

	result, err := runner.Run(context.Background(), SpecFilesList, CallOptions{})
	require.NoError(t, err)
	require.Equal(t, "/tmp/obsidian", executor.invocation.Binary)
	require.Equal(t, []string{"vault=Work", SpecFilesList.Name}, executor.invocation.Args)
	require.Equal(t, []string{"alpha", "beta"}, result.Parsed)
}

func TestRunnerWrapsMissingBinary(t *testing.T) {
	executor := &fakeExecutor{
		err: &osexec.Error{Name: "obsidian", Err: osexec.ErrNotFound},
	}
	runner := NewRunner(DefaultConfig(), executor)

	_, err := runner.Run(context.Background(), SpecVersion, CallOptions{})
	var binaryErr *BinaryNotFoundError
	require.ErrorAs(t, err, &binaryErr)
}

func TestRunnerWrapsParseErrors(t *testing.T) {
	executor := &fakeExecutor{
		result: ExecResult{Stdout: "{not-json}"},
	}
	runner := NewRunner(DefaultConfig(), executor)

	_, err := runner.Run(context.Background(), CommandSpec{Name: "broken", Output: OutputJSON}, CallOptions{})
	var parseErr *ParseError
	require.ErrorAs(t, err, &parseErr)
}

func TestRunnerWrapsCommandErrors(t *testing.T) {
	executor := &fakeExecutor{
		result: ExecResult{Stdout: "nope", Stderr: "boom", ExitCode: 23},
		err:    errors.New("failed"),
	}
	runner := NewRunner(DefaultConfig(), executor)

	_, err := runner.Run(context.Background(), SpecVersion, CallOptions{})
	var commandErr *CommandError
	require.ErrorAs(t, err, &commandErr)
	require.Equal(t, 23, commandErr.ExitCode)
}
