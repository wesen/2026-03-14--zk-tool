package obsidianmd

import (
	"strings"
)

// ExtractTasks returns markdown task text lines without the leading checkbox marker.
func ExtractTasks(content string) []string {
	lines := strings.Split(content, "\n")
	ret := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- [") || len(line) < 6 {
			continue
		}
		text := strings.TrimSpace(line[5:])
		if text == "" {
			continue
		}
		ret = append(ret, text)
	}
	return ret
}
