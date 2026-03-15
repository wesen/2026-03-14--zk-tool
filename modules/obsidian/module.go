package obsidianmod

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
	obsidianpkg "github.com/wesen/2026-03-14--zk-tool/pkg/obsidian"
	"github.com/wesen/2026-03-14--zk-tool/pkg/obsidiancli"
	"github.com/wesen/2026-03-14--zk-tool/pkg/obsidianmd"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
	"github.com/pkg/errors"
)

// Options customize runtime-state construction for the module.
type Options struct {
	NewRunner func(cfg obsidiancli.Config) obsidianpkg.Runner
	NewClient func(cfg obsidianpkg.Config, runner obsidianpkg.Runner) *obsidianpkg.Client
	NewOwner  func(vm *goja.Runtime) runtimeowner.Runner
}

// Module adapts the high-level Obsidian client into a goja native module.
type Module struct {
	opts   Options
	states sync.Map // map[*goja.Runtime]*runtimeState
}

type runtimeState struct {
	mu     sync.RWMutex
	cfg    obsidiancli.Config
	runner obsidianpkg.Runner
	client *obsidianpkg.Client
	owner  runtimeowner.Runner
}

var _ modules.NativeModule = (*Module)(nil)

// New creates a new native Obsidian module instance.
func New(opts Options) *Module {
	return &Module{opts: opts}
}

// Name returns the module name exposed to require().
func (m *Module) Name() string { return "obsidian" }

// Doc returns a short module description.
func (m *Module) Doc() string {
	return "Obsidian module with Promise-based vault operations, fluent queries, and markdown helpers."
}

// Loader wires the module exports for one runtime instance.
func (m *Module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)
	state := m.ensureState(vm)

	modules.SetExport(exports, m.Name(), "configure", func(call goja.FunctionCall) goja.Value {
		options, err := mapArg(vm, call.Argument(0))
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		cfg := mergeCLIConfig(state.config(), options)
		state.rebuild(cfg, m.opts)
		return vm.ToValue(configToJSMap(cfg))
	})

	modules.SetExport(exports, m.Name(), "version", func(goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "version", func(ctx context.Context, current *runtimeState) (any, error) {
			return current.clientSnapshot().Version(ctx)
		})
	})

	modules.SetExport(exports, m.Name(), "files", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "files", func(ctx context.Context, current *runtimeState) (any, error) {
			options, err := mapArg(vm, call.Argument(0))
			if err != nil {
				return nil, err
			}
			return current.clientSnapshot().Files(ctx, fileListOptions(options))
		})
	})

	modules.SetExport(exports, m.Name(), "read", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "read", func(ctx context.Context, current *runtimeState) (any, error) {
			ref := strings.TrimSpace(call.Argument(0).String())
			return current.clientSnapshot().Read(ctx, ref)
		})
	})

	modules.SetExport(exports, m.Name(), "create", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "create", func(ctx context.Context, current *runtimeState) (any, error) {
			title := strings.TrimSpace(call.Argument(0).String())
			options, err := mapArg(vm, call.Argument(1))
			if err != nil {
				return nil, err
			}
			return current.clientSnapshot().Create(ctx, title, createOptions(options))
		})
	})

	modules.SetExport(exports, m.Name(), "append", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "append", func(ctx context.Context, current *runtimeState) (any, error) {
			if err := current.clientSnapshot().Append(ctx, call.Argument(0).String(), call.Argument(1).String()); err != nil {
				return nil, err
			}
			return true, nil
		})
	})

	modules.SetExport(exports, m.Name(), "prepend", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "prepend", func(ctx context.Context, current *runtimeState) (any, error) {
			if err := current.clientSnapshot().Prepend(ctx, call.Argument(0).String(), call.Argument(1).String()); err != nil {
				return nil, err
			}
			return true, nil
		})
	})

	modules.SetExport(exports, m.Name(), "move", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "move", func(ctx context.Context, current *runtimeState) (any, error) {
			if err := current.clientSnapshot().Move(ctx, call.Argument(0).String(), call.Argument(1).String()); err != nil {
				return nil, err
			}
			return true, nil
		})
	})

	modules.SetExport(exports, m.Name(), "rename", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "rename", func(ctx context.Context, current *runtimeState) (any, error) {
			if err := current.clientSnapshot().Rename(ctx, call.Argument(0).String(), call.Argument(1).String()); err != nil {
				return nil, err
			}
			return true, nil
		})
	})

	modules.SetExport(exports, m.Name(), "delete", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "delete", func(ctx context.Context, current *runtimeState) (any, error) {
			options, err := mapArg(vm, call.Argument(1))
			if err != nil {
				return nil, err
			}
			if err := current.clientSnapshot().Delete(ctx, call.Argument(0).String(), deleteOptions(options)); err != nil {
				return nil, err
			}
			return true, nil
		})
	})

	modules.SetExport(exports, m.Name(), "note", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "note", func(ctx context.Context, current *runtimeState) (any, error) {
			note, err := current.clientSnapshot().Note(ctx, call.Argument(0).String())
			if err != nil {
				return nil, err
			}
			return noteToMap(ctx, note)
		})
	})

	modules.SetExport(exports, m.Name(), "query", func(call goja.FunctionCall) goja.Value {
		query := state.clientSnapshot().Query()
		options, err := mapArg(vm, call.Argument(0))
		if err == nil {
			applyQueryOptions(query, options)
		}
		return m.newQueryObject(vm, state, query)
	})

	modules.SetExport(exports, m.Name(), "batch", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "batch", func(ctx context.Context, current *runtimeState) (any, error) {
			options, err := mapArg(vm, call.Argument(0))
			if err != nil {
				return nil, err
			}
			query := current.clientSnapshot().Query()
			applyQueryOptions(query, options)

			var mapper goja.Callable
			if call.Argument(1) != nil && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
				var ok bool
				mapper, ok = goja.AssertFunction(call.Argument(1))
				if !ok {
					return nil, errors.New("obsidian module: batch mapper must be a function")
				}
			}

			notes, err := query.Run(ctx)
			if err != nil {
				return nil, err
			}

			results := make([]any, 0, len(notes))
			for _, note := range notes {
				noteValue, err := noteToMap(ctx, note)
				if err != nil {
					return nil, err
				}
				if mapper == nil {
					results = append(results, noteValue)
					continue
				}
				if current.owner != nil {
					mapped, err := current.owner.Call(ctx, "obsidian.batch.mapper", func(_ context.Context, vm *goja.Runtime) (any, error) {
						value, err := mapper(goja.Undefined(), vm.ToValue(noteValue))
						if err != nil {
							return nil, err
						}
						return value.Export(), nil
					})
					if err != nil {
						return nil, err
					}
					results = append(results, mapped)
					continue
				}
				value, err := mapper(goja.Undefined(), vm.ToValue(noteValue))
				if err != nil {
					return nil, err
				}
				results = append(results, value.Export())
			}
			return results, nil
		})
	})

	modules.SetExport(exports, m.Name(), "exec", func(call goja.FunctionCall) goja.Value {
		return m.promise(vm, state, "exec", func(ctx context.Context, current *runtimeState) (any, error) {
			name := strings.TrimSpace(call.Argument(0).String())
			if name == "" {
				return nil, errors.New("obsidian module: command name is empty")
			}
			parameters, err := mapArg(vm, call.Argument(1))
			if err != nil {
				return nil, err
			}
			flags, err := stringSliceArg(vm, call.Argument(2))
			if err != nil {
				return nil, err
			}
			result, err := current.runnerSnapshot().Run(ctx, obsidiancli.CommandSpec{
				Name:   name,
				Output: obsidiancli.OutputRaw,
			}, obsidiancli.CallOptions{
				Parameters: parameters,
				Flags:      flags,
			})
			if err != nil {
				return nil, err
			}
			return result.Stdout, nil
		})
	})

	if err := exports.Set("md", m.newMarkdownObject(vm)); err != nil {
		panic(vm.NewTypeError(err.Error()))
	}
}

func (m *Module) ensureState(vm *goja.Runtime) *runtimeState {
	if state, ok := m.states.Load(vm); ok {
		return state.(*runtimeState)
	}

	cfg := obsidiancli.DefaultConfig()
	state := &runtimeState{
		cfg: cfg,
	}
	state.rebuild(cfg, m.opts)

	if owner := m.opts.NewOwner; owner != nil {
		state.owner = owner(vm)
	}

	actual, _ := m.states.LoadOrStore(vm, state)
	return actual.(*runtimeState)
}

func (s *runtimeState) config() obsidiancli.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *runtimeState) runnerSnapshot() obsidianpkg.Runner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runner
}

func (s *runtimeState) clientSnapshot() *obsidianpkg.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client
}

func (s *runtimeState) rebuild(cfg obsidiancli.Config, opts Options) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cfg = cfg
	if opts.NewRunner != nil {
		s.runner = opts.NewRunner(cfg)
	} else {
		s.runner = obsidiancli.NewRunner(cfg, nil)
	}
	if opts.NewClient != nil {
		s.client = opts.NewClient(obsidianpkg.Config{CLI: cfg}, s.runner)
	} else {
		s.client = obsidianpkg.NewClient(obsidianpkg.Config{CLI: cfg}, s.runner)
	}
}

func (m *Module) promise(vm *goja.Runtime, state *runtimeState, label string, fn func(context.Context, *runtimeState) (any, error)) goja.Value {
	promise, resolve, reject := vm.NewPromise()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		value, err := fn(ctx, state)
		if state.owner != nil {
			_, _ = state.owner.Call(ctx, "obsidian.promise."+label, func(_ context.Context, vm *goja.Runtime) (any, error) {
				if err != nil {
					reject(vm.ToValue(err.Error()))
				} else {
					resolve(vm.ToValue(value))
				}
				return nil, nil
			})
			return
		}

		if err != nil {
			reject(vm.ToValue(err.Error()))
			return
		}
		resolve(vm.ToValue(value))
	}()

	return vm.ToValue(promise)
}

func mergeCLIConfig(cfg obsidiancli.Config, options map[string]any) obsidiancli.Config {
	for key, value := range options {
		switch strings.TrimSpace(strings.ToLower(key)) {
		case "binary", "binarypath":
			cfg.BinaryPath = strings.TrimSpace(toString(value))
		case "vault":
			cfg.Vault = strings.TrimSpace(toString(value))
		case "workingdir", "cwd":
			cfg.WorkingDir = strings.TrimSpace(toString(value))
		case "timeout", "timeoutms":
			cfg.Timeout = parseDurationValue(value)
		}
	}
	return cfg
}

func configToJSMap(cfg obsidiancli.Config) map[string]any {
	return map[string]any{
		"binaryPath": cfg.BinaryPath,
		"vault":      cfg.Vault,
		"workingDir": cfg.WorkingDir,
		"timeoutMs":  cfg.Timeout.Milliseconds(),
	}
}

func fileListOptions(options map[string]any) obsidianpkg.FileListOptions {
	return obsidianpkg.FileListOptions{
		Folder: strings.TrimSpace(toString(options["folder"])),
		Ext:    strings.TrimSpace(toString(options["ext"])),
		Limit:  toInt(options["limit"]),
		Vault:  strings.TrimSpace(toString(options["vault"])),
	}
}

func createOptions(options map[string]any) obsidianpkg.CreateOptions {
	return obsidianpkg.CreateOptions{
		Content:  toString(options["content"]),
		Folder:   strings.TrimSpace(toString(options["folder"])),
		Template: strings.TrimSpace(toString(options["template"])),
		Vault:    strings.TrimSpace(toString(options["vault"])),
	}
}

func deleteOptions(options map[string]any) obsidianpkg.DeleteOptions {
	return obsidianpkg.DeleteOptions{
		Permanent: toBool(options["permanent"]),
		Vault:     strings.TrimSpace(toString(options["vault"])),
	}
}

func applyQueryOptions(query *obsidianpkg.Query, options map[string]any) {
	if query == nil {
		return
	}
	if folder := strings.TrimSpace(toString(options["folder"])); folder != "" {
		query.InFolder(folder)
	}
	if ext := strings.TrimSpace(toString(options["ext"])); ext != "" {
		query.WithExtension(ext)
	}
	if tag := strings.TrimSpace(toString(options["tag"])); tag != "" {
		query.Tagged(tag)
	}
	if contains := strings.TrimSpace(toString(options["contains"])); contains != "" {
		query.Contains(contains)
	}
	if limit := toInt(options["limit"]); limit > 0 {
		query.Limit(limit)
	}
	if vault := strings.TrimSpace(toString(options["vault"])); vault != "" {
		query.InVault(vault)
	}
}

func (m *Module) newQueryObject(vm *goja.Runtime, state *runtimeState, query *obsidianpkg.Query) *goja.Object {
	obj := vm.NewObject()

	_ = obj.Set("inFolder", func(folder string) goja.Value {
		query.InFolder(folder)
		return obj
	})
	_ = obj.Set("withExtension", func(ext string) goja.Value {
		query.WithExtension(ext)
		return obj
	})
	_ = obj.Set("tagged", func(tag string) goja.Value {
		query.Tagged(tag)
		return obj
	})
	_ = obj.Set("contains", func(term string) goja.Value {
		query.Contains(term)
		return obj
	})
	_ = obj.Set("limit", func(limit int) goja.Value {
		query.Limit(limit)
		return obj
	})
	_ = obj.Set("inVault", func(vault string) goja.Value {
		query.InVault(vault)
		return obj
	})
	_ = obj.Set("run", func() goja.Value {
		return m.promise(vm, state, "query.run", func(ctx context.Context, current *runtimeState) (any, error) {
			notes, err := query.Run(ctx)
			if err != nil {
				return nil, err
			}
			ret := make([]map[string]any, 0, len(notes))
			for _, note := range notes {
				row, err := noteToMap(ctx, note)
				if err != nil {
					return nil, err
				}
				ret = append(ret, row)
			}
			return ret, nil
		})
	})

	return obj
}

func (m *Module) newMarkdownObject(vm *goja.Runtime) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("frontmatter", func(content string) map[string]any {
		ret, err := obsidianmd.ParseFrontmatter(content)
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		return ret
	})
	_ = obj.Set("tags", func(content string) []string {
		return obsidianmd.ExtractTags(content)
	})
	_ = obj.Set("wikilinks", func(content string) []string {
		return obsidianmd.ExtractWikilinks(content)
	})
	_ = obj.Set("headings", func(content string) []string {
		return obsidianmd.ExtractHeadings(content)
	})
	_ = obj.Set("tasks", func(content string) []string {
		return obsidianmd.ExtractTasks(content)
	})
	return obj
}

func noteToMap(ctx context.Context, note *obsidianpkg.Note) (map[string]any, error) {
	content, err := note.Content(ctx)
	if err != nil {
		return nil, err
	}
	tags, err := note.Tags(ctx)
	if err != nil {
		return nil, err
	}
	headings, err := note.Headings(ctx)
	if err != nil {
		return nil, err
	}
	tasks, err := note.Tasks(ctx)
	if err != nil {
		return nil, err
	}
	wikilinks, err := note.Wikilinks(ctx)
	if err != nil {
		return nil, err
	}
	frontmatter, err := note.Frontmatter(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"path":        note.Path(),
		"title":       note.Title(),
		"content":     content,
		"tags":        tags,
		"headings":    headings,
		"tasks":       tasks,
		"wikilinks":   wikilinks,
		"frontmatter": frontmatter,
	}, nil
}

func mapArg(vm *goja.Runtime, value goja.Value) (map[string]any, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return map[string]any{}, nil
	}
	exported := value.Export()
	switch v := exported.(type) {
	case map[string]any:
		return v, nil
	default:
		return nil, fmt.Errorf("expected object argument, got %T", exported)
	}
}

func stringSliceArg(vm *goja.Runtime, value goja.Value) ([]string, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, nil
	}
	exported := value.Export()
	switch v := exported.(type) {
	case []string:
		return v, nil
	case []any:
		ret := make([]string, 0, len(v))
		for _, item := range v {
			ret = append(ret, toString(item))
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("expected string array, got %T", exported)
	}
}

func toString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func toInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return 0
	}
}

func toBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	default:
		return false
	}
}

func parseDurationValue(value any) time.Duration {
	switch v := value.(type) {
	case time.Duration:
		return v
	case int:
		return time.Duration(v) * time.Millisecond
	case int64:
		return time.Duration(v) * time.Millisecond
	case float64:
		return time.Duration(v) * time.Millisecond
	case string:
		duration, err := time.ParseDuration(v)
		if err == nil {
			return duration
		}
	}
	return 0
}

func init() {
	modules.Register(New(Options{}))
}
