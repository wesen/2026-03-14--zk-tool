package obsidian

import (
	"context"
	"testing"

	"github.com/wesen/2026-03-14--zk-tool/pkg/obsidiancli"
	"github.com/stretchr/testify/require"
)

type fakeRunner struct {
	t        *testing.T
	handlers map[string]func(call obsidiancli.CallOptions) (obsidiancli.Result, error)
}

func (f *fakeRunner) Run(_ context.Context, spec obsidiancli.CommandSpec, call obsidiancli.CallOptions) (obsidiancli.Result, error) {
	handler, ok := f.handlers[spec.Name]
	require.Truef(f.t, ok, "unexpected command: %s", spec.Name)
	return handler(call)
}

func TestClientVersion(t *testing.T) {
	client := NewClient(Config{}, &fakeRunner{
		t: t,
		handlers: map[string]func(call obsidiancli.CallOptions) (obsidiancli.Result, error){
			obsidiancli.SpecVersion.Name: func(call obsidiancli.CallOptions) (obsidiancli.Result, error) {
				require.Empty(t, call.Parameters)
				return obsidiancli.Result{Parsed: "1.12.4", Stdout: "1.12.4\n"}, nil
			},
		},
	})

	version, err := client.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, "1.12.4", version)
}

func TestClientReadResolvesWikilinkStyleReferences(t *testing.T) {
	client := NewClient(Config{}, &fakeRunner{
		t: t,
		handlers: map[string]func(call obsidiancli.CallOptions) (obsidiancli.Result, error){
			obsidiancli.SpecFilesList.Name: func(call obsidiancli.CallOptions) (obsidiancli.Result, error) {
				return obsidiancli.Result{Parsed: []string{"Notes/Systems.md"}}, nil
			},
			obsidiancli.SpecFileRead.Name: func(call obsidiancli.CallOptions) (obsidiancli.Result, error) {
				require.Equal(t, "Notes/Systems.md", call.Parameters["path"])
				return obsidiancli.Result{Parsed: "# Systems"}, nil
			},
		},
	})

	content, err := client.Read(context.Background(), "Systems")
	require.NoError(t, err)
	require.Equal(t, "# Systems", content)
}

func TestClientQueryFiltersByTag(t *testing.T) {
	client := NewClient(Config{}, &fakeRunner{
		t: t,
		handlers: map[string]func(call obsidiancli.CallOptions) (obsidiancli.Result, error){
			obsidiancli.SpecFilesList.Name: func(call obsidiancli.CallOptions) (obsidiancli.Result, error) {
				require.Equal(t, "ZK", call.Parameters["folder"])
				return obsidiancli.Result{Parsed: []string{"ZK/A.md", "ZK/B.md"}}, nil
			},
			obsidiancli.SpecFileRead.Name: func(call obsidiancli.CallOptions) (obsidiancli.Result, error) {
				path := call.Parameters["path"].(string)
				switch path {
				case "ZK/A.md":
					return obsidiancli.Result{Parsed: "# A\n\n#system\n"}, nil
				case "ZK/B.md":
					return obsidiancli.Result{Parsed: "# B\n\n#draft\n"}, nil
				default:
					t.Fatalf("unexpected path %s", path)
					return obsidiancli.Result{}, nil
				}
			},
		},
	})

	notes, err := client.Query().
		InFolder("ZK").
		Tagged("system").
		Run(context.Background())
	require.NoError(t, err)
	require.Len(t, notes, 1)
	require.Equal(t, "ZK/A.md", notes[0].Path())
}
