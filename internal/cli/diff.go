package cli

import (
	"fmt"
	"strings"
)

func Diff(existing, incoming string) string {
	existingMap := parseEnv(existing)
	incomingMap := parseEnv(incoming)

	var changes []string

	for key, newVal := range incomingMap {
		oldVal, exists := existingMap[key]
		if !exists {
			changes = append(changes, fmt.Sprintf("  + %s=*** (new)", key))
		} else if oldVal != newVal {
			changes = append(changes, fmt.Sprintf("  ~ %s=*** (changed)", key))
		}
	}

	for key := range existingMap {
		if _, exists := incomingMap[key]; !exists {
			changes = append(changes, fmt.Sprintf("  - %s (removed)", key))
		}
	}

	if len(changes) == 0 {
		return ""
	}

	return "Changes:\n" + strings.Join(changes, "\n")
}

func parseEnv(content string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}
