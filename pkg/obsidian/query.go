package obsidian

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

type queryMode string

const (
	queryModeFiles  queryMode = "files"
	queryModeSearch queryMode = "search"
)

// Query is a fluent client-side note selector built on top of CLI primitives.
type Query struct {
	client *Client
	mode   queryMode

	folder   string
	ext      string
	tag      string
	contains string
	limit    int
	vault    string
}

// InFolder narrows the query to one folder prefix.
func (q *Query) InFolder(folder string) *Query {
	q.folder = strings.TrimSpace(folder)
	return q
}

// WithExtension narrows the query to one file extension.
func (q *Query) WithExtension(ext string) *Query {
	q.ext = normalizeExt(ext)
	return q
}

// Tagged filters note contents by hashtag after candidate expansion.
func (q *Query) Tagged(tag string) *Query {
	q.tag = strings.TrimSpace(strings.TrimPrefix(tag, "#"))
	return q
}

// Contains switches the primary mode to text search.
func (q *Query) Contains(term string) *Query {
	q.contains = strings.TrimSpace(term)
	if q.contains != "" {
		q.mode = queryModeSearch
	}
	return q
}

// Limit truncates the final note set.
func (q *Query) Limit(limit int) *Query {
	if limit > 0 {
		q.limit = limit
	}
	return q
}

// InVault targets one vault name for the underlying CLI calls.
func (q *Query) InVault(vault string) *Query {
	q.vault = strings.TrimSpace(vault)
	return q
}

// Run resolves the current query into note objects.
func (q *Query) Run(ctx context.Context) ([]*Note, error) {
	if q == nil || q.client == nil {
		return nil, errors.New("obsidian: query has no client")
	}

	var (
		paths []string
		err   error
	)

	switch q.mode {
	case queryModeSearch:
		if q.contains == "" {
			return nil, errors.New("obsidian: contains term is empty")
		}
		paths, err = q.client.Search(ctx, q.contains, SearchOptions{
			Folder: q.folder,
			Ext:    q.ext,
			Limit:  q.limit,
			Vault:  q.vault,
		})
	default:
		paths, err = q.client.Files(ctx, FileListOptions{
			Folder: q.folder,
			Ext:    q.ext,
			Limit:  q.limit,
			Vault:  q.vault,
		})
	}
	if err != nil {
		return nil, err
	}

	notes := make([]*Note, 0, len(paths))
	for _, path := range paths {
		notes = append(notes, q.client.noteForPath(path))
	}

	if q.tag != "" {
		filtered := make([]*Note, 0, len(notes))
		for _, note := range notes {
			tags, tagsErr := note.Tags(ctx)
			if tagsErr != nil {
				return nil, tagsErr
			}
			if containsStringFold(tags, q.tag) {
				filtered = append(filtered, note)
			}
		}
		notes = filtered
	}

	if q.contains != "" && q.mode != queryModeSearch {
		filtered := make([]*Note, 0, len(notes))
		for _, note := range notes {
			content, readErr := note.Content(ctx)
			if readErr != nil {
				return nil, readErr
			}
			if strings.Contains(strings.ToLower(content), strings.ToLower(q.contains)) {
				filtered = append(filtered, note)
			}
		}
		notes = filtered
	}

	if q.limit > 0 && len(notes) > q.limit {
		notes = notes[:q.limit]
	}

	return notes, nil
}

func containsStringFold(values []string, needle string) bool {
	needle = strings.TrimSpace(strings.ToLower(needle))
	for _, value := range values {
		if strings.TrimSpace(strings.ToLower(value)) == needle {
			return true
		}
	}
	return false
}
