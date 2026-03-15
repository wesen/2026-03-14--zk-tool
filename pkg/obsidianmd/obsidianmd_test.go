package obsidianmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFrontmatter(t *testing.T) {
	content := `---
title: Example
tags:
  - zk
  - notes
---

# Example
`

	meta, err := ParseFrontmatter(content)
	require.NoError(t, err)
	require.Equal(t, "Example", meta["title"])
}

func TestExtractors(t *testing.T) {
	content := `---
tags:
  - one
---

# Heading

Text with #two and [[Third Note]].

- [ ] Task
`

	require.Equal(t, []string{"Heading"}, ExtractHeadings(content))
	require.Equal(t, []string{"two"}, ExtractTags(content))
	require.Equal(t, []string{"Task"}, ExtractTasks(content))
	require.Equal(t, []string{"Third Note"}, ExtractWikilinks(content))
}

func TestNoteBuilder(t *testing.T) {
	note, err := NewNoteBuilder("Example").
		WithFrontmatter("status", "draft").
		WithTags("one", "#two").
		WithBody("Paragraph 1", "Paragraph 2").
		Build()
	require.NoError(t, err)
	require.Contains(t, note, "status: draft")
	require.Contains(t, note, "# Example")
	require.Contains(t, note, "Paragraph 1")
	require.Contains(t, note, "tags:")
}
