package obsidianjs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	_ "github.com/wesen/2026-03-14--zk-tool/modules/obsidian"
	"github.com/pkg/errors"
)

// RunOptions configures one JavaScript script execution.
type RunOptions struct {
	ScriptPath string
	BinaryPath string
	Vault      string
}

// RunResult captures the evaluated script result.
type RunResult struct {
	ScriptPath string
	BinaryPath string
	Vault      string
	Output     string
	Parsed     any
}

// RunFile executes one JavaScript file with the local obsidian module pre-registered.
func RunFile(ctx context.Context, opts RunOptions) (*RunResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	scriptPath, err := filepath.Abs(strings.TrimSpace(opts.ScriptPath))
	if err != nil {
		return nil, errors.Wrap(err, "resolve script path")
	}
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, errors.Wrap(err, "read script file")
	}

	factory, err := engine.NewBuilder(
		engine.WithModuleRootsFromScript(scriptPath, engine.DefaultModuleRootsOptions()),
	).WithModules(
		engine.DefaultRegistryModules(),
	).Build()
	if err != nil {
		return nil, errors.Wrap(err, "build goja runtime")
	}

	rt, err := factory.NewRuntime(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "create goja runtime")
	}
	defer func() {
		_ = rt.Close(context.Background())
	}()

	code, err := bootstrapCode(string(raw), opts)
	if err != nil {
		return nil, err
	}

	valueAny, err := rt.Owner.Call(ctx, "obsidianjs.run", func(_ context.Context, vm *goja.Runtime) (any, error) {
		return vm.RunString(code)
	})
	if err != nil {
		return nil, errors.Wrap(err, "execute script")
	}

	value, ok := valueAny.(goja.Value)
	if !ok {
		return nil, errors.Errorf("unexpected result type %T", valueAny)
	}

	output, err := stringifyValue(ctx, rt, value)
	if err != nil {
		return nil, errors.Wrap(err, "settle script result")
	}

	result := &RunResult{
		ScriptPath: scriptPath,
		BinaryPath: resolvedBinaryPath(opts.BinaryPath),
		Vault:      strings.TrimSpace(opts.Vault),
		Output:     output,
		Parsed:     parseJSONOutput(output),
	}
	if result.Parsed == nil {
		result.Parsed = output
	}

	return result, nil
}

func bootstrapCode(script string, opts RunOptions) (string, error) {
	config := map[string]any{
		"binaryPath": resolvedBinaryPath(opts.BinaryPath),
	}
	if vault := strings.TrimSpace(opts.Vault); vault != "" {
		config["vault"] = vault
	}
	rawConfig, err := json.Marshal(config)
	if err != nil {
		return "", errors.Wrap(err, "marshal obsidian config")
	}

	var builder strings.Builder
	builder.WriteString("require(\"obsidian\").configure(")
	builder.Write(rawConfig)
	builder.WriteString(");\n")
	builder.WriteString(script)
	return builder.String(), nil
}

func resolvedBinaryPath(configured string) string {
	if configured = strings.TrimSpace(configured); configured != "" {
		return configured
	}
	home, err := os.UserHomeDir()
	if err == nil {
		candidate := filepath.Join(home, ".local", "bin", "obsidian")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate
		}
	}
	return "obsidian"
}

func stringifyValue(ctx context.Context, rt *engine.Runtime, value goja.Value) (string, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return "", nil
	}
	if promise, ok := value.Export().(*goja.Promise); ok {
		return waitForPromise(ctx, rt, promise)
	}
	return value.String(), nil
}

func waitForPromise(ctx context.Context, rt *engine.Runtime, promise *goja.Promise) (string, error) {
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		snapshotAny, err := rt.Owner.Call(ctx, "obsidianjs.promise", func(_ context.Context, vm *goja.Runtime) (any, error) {
			return promiseSnapshot{
				State:  promise.State(),
				Result: promise.Result(),
			}, nil
		})
		if err != nil {
			return "", err
		}

		snapshot, ok := snapshotAny.(promiseSnapshot)
		if !ok {
			return "", errors.Errorf("unexpected promise snapshot type %T", snapshotAny)
		}

		switch snapshot.State {
		case goja.PromiseStatePending:
			time.Sleep(5 * time.Millisecond)
		case goja.PromiseStateRejected:
			return "", errors.Errorf("Promise rejected: %s", valueString(snapshot.Result))
		case goja.PromiseStateFulfilled:
			return valueString(snapshot.Result), nil
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

type promiseSnapshot struct {
	State  goja.PromiseState
	Result goja.Value
}

func valueString(value goja.Value) string {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return ""
	}
	return value.String()
}

func parseJSONOutput(output string) any {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}
	if !strings.HasPrefix(output, "{") && !strings.HasPrefix(output, "[") {
		return nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return nil
	}
	return parsed
}
