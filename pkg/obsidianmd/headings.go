package obsidianmd

import (
	"strings"
)

// ExtractHeadings returns markdown ATX heading text in document order.
func ExtractHeadings(content string) []string {
	lines := strings.Split(content, "\n")
	ret := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		text := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if text == "" {
			continue
		}
		ret = append(ret, text)
	}
	return ret
}
