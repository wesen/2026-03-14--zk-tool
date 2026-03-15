package obsidianmd

import (
	"bytes"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// NoteBuilder assists with assembling Obsidian-flavored markdown notes.
type NoteBuilder struct {
	title       string
	frontmatter map[string]any
	body        []string
}

// NewNoteBuilder creates a new note builder.
func NewNoteBuilder(title string) *NoteBuilder {
	return &NoteBuilder{
		title:       strings.TrimSpace(title),
		frontmatter: map[string]any{},
	}
}

// WithFrontmatter sets one frontmatter key.
func (b *NoteBuilder) WithFrontmatter(key string, value any) *NoteBuilder {
	key = strings.TrimSpace(key)
	if key == "" {
		return b
	}
	b.frontmatter[key] = value
	return b
}

// WithBody appends raw markdown body sections.
func (b *NoteBuilder) WithBody(parts ...string) *NoteBuilder {
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		b.body = append(b.body, part)
	}
	return b
}

// WithTags stores tags in frontmatter.
func (b *NoteBuilder) WithTags(tags ...string) *NoteBuilder {
	if len(tags) == 0 {
		return b
	}
	filtered := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
		if tag == "" {
			continue
		}
		filtered = append(filtered, tag)
	}
	if len(filtered) == 0 {
		return b
	}
	sort.Strings(filtered)
	b.frontmatter["tags"] = filtered
	return b
}

// Build renders the note into Obsidian markdown.
func (b *NoteBuilder) Build() (string, error) {
	if b == nil {
		return "", errors.New("obsidianmd: note builder is nil")
	}

	var out bytes.Buffer
	if len(b.frontmatter) > 0 {
		raw, err := marshalFrontmatter(b.frontmatter)
		if err != nil {
			return "", err
		}
		out.WriteString(raw)
		out.WriteString("\n")
	}

	if b.title != "" {
		out.WriteString("# ")
		out.WriteString(b.title)
		out.WriteString("\n\n")
	}

	for i, part := range b.body {
		if i > 0 {
			out.WriteString("\n\n")
		}
		out.WriteString(part)
	}

	return strings.TrimRight(out.String(), "\n") + "\n", nil
}

func marshalFrontmatter(data map[string]any) (string, error) {
	raw, err := yaml.Marshal(data)
	if err != nil {
		return "", errors.Wrap(err, "obsidianmd: marshal frontmatter")
	}
	return "---\n" + strings.TrimRight(string(raw), "\n") + "\n---", nil
}
