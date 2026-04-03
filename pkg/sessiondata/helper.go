package sessiondata

import (
	"fmt"
	"net/http"

	"httpDebugger/pkg/sortedMap"
)

func compareStringField(original, other string) *FieldDiff {
	return &FieldDiff{
		Original: original,
		Other:    other,
		Changed:  original != other,
	}
}

func compareHeaders(original, other *sortedMap.SortedMap) *HeadersDiff {
	diff := &HeadersDiff{
		Added:    make(map[string]string),
		Removed:  make(map[string]string),
		Modified: make(map[string]FieldDiff),
	}

	for _, key := range original.Keys() {
		originalVal, _ := original.Get(key)
		otherVal, otherExists := other.Get(key)
		if !otherExists {
			diff.Removed[key] = fmt.Sprintf("%v", originalVal)
			diff.Changed = true
		} else if fmt.Sprintf("%v", originalVal) != fmt.Sprintf("%v", otherVal) {
			diff.Modified[key] = FieldDiff{
				Original: fmt.Sprintf("%v", originalVal),
				Other:    fmt.Sprintf("%v", otherVal),
				Changed:  true,
			}
			diff.Changed = true
		}
	}

	for _, key := range other.Keys() {
		if _, exists := original.Get(key); !exists {
			otherVal, _ := other.Get(key)
			diff.Added[key] = fmt.Sprintf("%v", otherVal)
			diff.Changed = true
		}
	}

	return diff
}

func compareCookies(original, other map[string]string) *CookiesDiff {
	diff := &CookiesDiff{
		Added:    make(map[string]string),
		Removed:  make(map[string]string),
		Modified: make(map[string]FieldDiff),
	}

	for key, originalVal := range original {
		otherVal, otherExists := other[key]
		if !otherExists {
			diff.Removed[key] = originalVal
			diff.Changed = true
		} else if originalVal != otherVal {
			diff.Modified[key] = FieldDiff{
				Original: originalVal,
				Other:    otherVal,
				Changed:  true,
			}
			diff.Changed = true
		}
	}

	for key, otherVal := range other {
		if _, exists := original[key]; !exists {
			diff.Added[key] = otherVal
			diff.Changed = true
		}
	}

	return diff
}

func isWsUpgradeRequest(r *http.Request) bool {
	return r.Method == http.MethodGet &&
		r.Header.Get("Upgrade") == "websocket" &&
		r.Header.Get("Sec-WebSocket-Version") == "13" &&
		r.Header.Get("Sec-WebSocket-Key") != ""
}

func newMessageStats() MessageStats {
	return MessageStats{}
}
