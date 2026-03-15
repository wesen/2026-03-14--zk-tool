package obsidianmd

import (
	"regexp"
	"sort"
)

var wikilinkRe = regexp.MustCompile(`\[\[([^\]|#]+)(?:#[^\]|]+)?(?:\|[^\]]+)?\]\]`)

// ExtractWikilinks returns unique wikilink targets found in markdown content.
func ExtractWikilinks(content string) []string {
	matches := wikilinkRe.FindAllStringSubmatch(content, -1)
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		seen[match[1]] = struct{}{}
	}
	ret := make([]string, 0, len(seen))
	for link := range seen {
		ret = append(ret, link)
	}
	sort.Strings(ret)
	return ret
}
