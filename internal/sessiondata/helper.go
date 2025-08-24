package sessiondata

import (
	"fmt"
	"httpDebugger/internal/sortedMap"
	"net/http"
	"reflect"

	fhttp "github.com/bogdanfinn/fhttp"
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
		Added:    make(map[string]interface{}),
		Removed:  make(map[string]interface{}),
		Modified: make(map[string]FieldDiff),
	}

	originalKeys := original.Keys()
	otherKeys := other.Keys()

	allKeys := make(map[string]bool)
	for _, key := range originalKeys {
		allKeys[key] = true
	}
	for _, key := range otherKeys {
		allKeys[key] = true
	}

	for key := range allKeys {
		originalVal, originalExists := original.Get(key)
		otherVal, otherExists := other.Get(key)

		if originalExists && !otherExists {
			diff.Removed[key] = originalVal
			diff.Changed = true
		} else if !originalExists && otherExists {
			diff.Added[key] = otherVal
			diff.Changed = true
		} else if originalExists && otherExists {
			if !reflect.DeepEqual(originalVal, otherVal) {
				diff.Modified[key] = FieldDiff{
					Original: fmt.Sprintf("%v", originalVal),
					Other:    fmt.Sprintf("%v", otherVal),
					Changed:  true,
				}
				diff.Changed = true
			}
		}
	}

	return diff
}

func compareCookies(original, other map[string]string) *CookiesDiff {
	diff := &CookiesDiff{
		Added:    make(map[string]interface{}),
		Removed:  make(map[string]interface{}),
		Modified: make(map[string]FieldDiff),
	}

	allKeys := make(map[string]bool)
	for key := range original {
		allKeys[key] = true
	}
	for key := range other {
		allKeys[key] = true
	}

	for key := range allKeys {
		originalVal, originalExists := original[key]
		otherVal, otherExists := other[key]

		if originalExists && !otherExists {
			diff.Removed[key] = originalVal
			diff.Changed = true
		} else if !originalExists && otherExists {
			diff.Added[key] = otherVal
			diff.Changed = true
		} else if originalExists && otherExists {
			if !reflect.DeepEqual(originalVal, otherVal) {
				diff.Modified[key] = FieldDiff{
					Original: fmt.Sprintf("%v", originalVal),
					Other:    fmt.Sprintf("%v", otherVal),
					Changed:  true,
				}
				diff.Changed = true
			}
		}
	}

	return diff
}

func isWsUpgradeRequest(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}

	if r.Header.Get("Upgrade") != "websocket" {
		return false
	}

	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		return false
	}

	if r.Header.Get("Sec-WebSocket-Key") == "" {
		return false
	}

	return true
}

func newMessageStats() MessageStats {
	return MessageStats{
		TotalMessages:    0,
		InboundMessages:  0,
		OutboundMessages: 0,
		TextMessages:     0,
		BinaryMessages:   0,
		ControlFrames:    0,
		TotalBytes:       0,
		InboundBytes:     0,
		OutboundBytes:    0,
	}
}

func sortedMapToHeaders(m *sortedMap.SortedMap) fhttp.Header {
	headers := fhttp.Header{}
	for _, key := range m.Keys() {
		val, _ := m.Get(key)
		if val != nil {
			headers.Add(key, fmt.Sprintf("%v", val))
		}
	}
	headers["Header-Order:"] = m.Order

	return headers
}
