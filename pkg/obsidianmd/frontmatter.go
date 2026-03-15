package obsidianmd

import (
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/pkg/errors"
)

// ParseFrontmatter extracts YAML frontmatter into a map.
func ParseFrontmatter(content string) (map[string]any, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return map[string]any{}, nil
	}

	data := map[string]any{}
	rest, err := frontmatter.Parse(strings.NewReader(content), &data)
	if err != nil {
		if strings.Contains(err.Error(), "frontmatter delimiter not found") {
			return map[string]any{}, nil
		}
		return nil, errors.Wrap(err, "obsidianmd: parse frontmatter")
	}
	if rest == nil {
		return map[string]any{}, nil
	}
	if len(data) == 0 {
		return map[string]any{}, nil
	}
	return data, nil
}
