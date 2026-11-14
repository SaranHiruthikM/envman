package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

func Diff(existing, incoming string) string {
	existingMap, _ := godotenv.Unmarshal(existing)
	incomingMap, _ := godotenv.Unmarshal(incoming)

	if existingMap == nil {
		existingMap = map[string]string{}
	}
	if incomingMap == nil {
		incomingMap = map[string]string{}
	}

	var added, changed, removed []string
	for key, newVal := range incomingMap {
		oldVal, exists := existingMap[key]
		switch {
		case !exists:
			added = append(added, fmt.Sprintf("  + %s=*** (new)", key))
		case oldVal != newVal:
			changed = append(changed, fmt.Sprintf("  ~ %s=*** (changed)", key))
		}
	}
	for key := range existingMap {
		if _, exists := incomingMap[key]; !exists {
			removed = append(removed, fmt.Sprintf("  - %s (removed)", key))
		}
	}

	sort.Strings(added)
	sort.Strings(changed)
	sort.Strings(removed)

	all := append(added, changed...)
	all = append(all, removed...)
	if len(all) == 0 {
		return ""
	}
	return "Changes:\n" + strings.Join(all, "\n")
}
