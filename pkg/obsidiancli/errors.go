package obsidiancli

import (
	"fmt"
	"strings"
)

// BinaryNotFoundError indicates the configured CLI binary could not be found.
type BinaryNotFoundError struct {
	Path string
}

func (e *BinaryNotFoundError) Error() string {
	if e == nil {
		return "obsidiancli: binary not found"
	}
	if strings.TrimSpace(e.Path) == "" {
		return "obsidiancli: binary not found"
	}
	return fmt.Sprintf("obsidiancli: binary not found: %s", e.Path)
}

// CommandError captures a non-successful command execution.
type CommandError struct {
	Spec     CommandSpec
	Args     []string
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func (e *CommandError) Error() string {
	if e == nil {
		return "obsidiancli: command failed"
	}
	if e.Err == nil {
		return fmt.Sprintf("obsidiancli: command %s failed", e.Spec.Name)
	}
	return fmt.Sprintf("obsidiancli: command %s failed: %v", e.Spec.Name, e.Err)
}

func (e *CommandError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ParseError indicates stdout could not be parsed according to the declared output format.
type ParseError struct {
	Spec   CommandSpec
	Stdout string
	Err    error
}

func (e *ParseError) Error() string {
	if e == nil {
		return "obsidiancli: parse error"
	}
	if e.Err == nil {
		return fmt.Sprintf("obsidiancli: parse %s output", e.Spec.Name)
	}
	return fmt.Sprintf("obsidiancli: parse %s output: %v", e.Spec.Name, e.Err)
}

func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
