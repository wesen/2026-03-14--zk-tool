package obsidian

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/wesen/2026-03-14--zk-tool/pkg/obsidianmd"
)

// Note provides a lazily hydrated view over one note path.
type Note struct {
	client  *Client
	path    string
	title   string
	content string
	loaded  bool
}

func (c *Client) noteForPath(path string) *Note {
	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return &Note{
		client: c,
		path:   path,
		title:  title,
	}
}

// Path returns the canonical vault path for this note.
func (n *Note) Path() string { return n.path }

// Title returns the current note title.
func (n *Note) Title() string { return n.title }

// Content loads the note body if necessary and returns it.
func (n *Note) Content(ctx context.Context) (string, error) {
	if err := n.ensureLoaded(ctx); err != nil {
		return "", err
	}
	return n.content, nil
}

// Frontmatter parses the YAML frontmatter block if present.
func (n *Note) Frontmatter(ctx context.Context) (map[string]any, error) {
	content, err := n.Content(ctx)
	if err != nil {
		return nil, err
	}
	return obsidianmd.ParseFrontmatter(content)
}

// Tags returns hashtag-style tags discovered in the note body and frontmatter.
func (n *Note) Tags(ctx context.Context) ([]string, error) {
	content, err := n.Content(ctx)
	if err != nil {
		return nil, err
	}
	return obsidianmd.ExtractTags(content), nil
}

// Wikilinks returns wikilinks referenced in the note body.
func (n *Note) Wikilinks(ctx context.Context) ([]string, error) {
	content, err := n.Content(ctx)
	if err != nil {
		return nil, err
	}
	return obsidianmd.ExtractWikilinks(content), nil
}

// Headings returns heading texts in document order.
func (n *Note) Headings(ctx context.Context) ([]string, error) {
	content, err := n.Content(ctx)
	if err != nil {
		return nil, err
	}
	return obsidianmd.ExtractHeadings(content), nil
}

// Tasks returns markdown task lines from the note.
func (n *Note) Tasks(ctx context.Context) ([]string, error) {
	content, err := n.Content(ctx)
	if err != nil {
		return nil, err
	}
	return obsidianmd.ExtractTasks(content), nil
}

// Reload discards any cached content for this note.
func (n *Note) Reload() {
	n.loaded = false
	n.content = ""
	if n.client != nil {
		n.client.Invalidate(n.path)
	}
}

func (n *Note) ensureLoaded(ctx context.Context) error {
	if n.loaded {
		return nil
	}
	content, err := n.client.Read(ctx, n.path)
	if err != nil {
		return err
	}
	n.content = content
	n.loaded = true
	return nil
}
