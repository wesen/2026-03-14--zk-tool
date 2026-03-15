package obsidianmd

import (
	"regexp"
	"sort"
)

var tagRe = regexp.MustCompile(`(^|[[:space:][:punct:]])#([[:alnum:]_/-]+)`)

// ExtractTags finds hashtag-style tags from markdown content.
func ExtractTags(content string) []string {
	matches := tagRe.FindAllStringSubmatch(content, -1)
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		seen[match[2]] = struct{}{}
	}
	ret := make([]string, 0, len(seen))
	for tag := range seen {
		ret = append(ret, tag)
	}
	sort.Strings(ret)
	return ret
}
