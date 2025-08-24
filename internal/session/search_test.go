package session

import (
	"httpDebugger/internal/sessiondata"
	"httpDebugger/internal/sortedMap"
	"sort"
	"testing"
)

func setupTestStoreWithData() *InMemoryStore {
	store := NewInMemoryStore(100)

	sessions := []*sessiondata.Session{
		{
			ID: "1",
			Request: &sessiondata.RequestData{
				URL:  "https://www.amazon.com/products/search?q=laptop",
				Body: "search query for laptop computers",
				Headers: sortedMap.SortedMap{
					Entries: map[string]interface{}{
						"Content-Type":  "application/json",
						"Authorization": "Bearer token123",
						"User-Agent":    "Mozilla/5.0 Chrome",
						"X-Custom":      "value1",
					},
					Order: []string{"Content-Type", "Authorization", "User-Agent", "X-Custom"},
				},
				Cookies: map[string]interface{}{
					"session_id": "abc123",
					"user_pref":  "theme=dark",
					"cart_items": 5,
				},
			},
			Response: &sessiondata.ResponseData{
				Body: "laptop search results with 100 products",
			},
		},
		{
			ID: "2",
			Request: &sessiondata.RequestData{
				URL:  "https://www.google.com/search?q=golang+tutorial",
				Body: "golang tutorial request data",
				Headers: sortedMap.SortedMap{
					Entries: map[string]interface{}{
						"Content-Type": "text/html",
						"Accept":       "text/html,application/xml",
						"Cookie":       "CONSENT=YES+cb.20210328-17-p0.en+FX+667",
					},
					Order: []string{"Content-Type", "Accept", "Cookie"},
				},
				Cookies: map[string]interface{}{
					"search_pref": "safe=on",
					"lang":        "en",
					"user_id":     12345,
				},
			},
			Response: &sessiondata.ResponseData{
				Body: "golang tutorial results and documentation",
			},
		},
		{
			ID: "3",
			Request: &sessiondata.RequestData{
				URL:  "https://api.github.com/repos/user/project",
				Body: "",
				Headers: sortedMap.SortedMap{
					Entries: map[string]interface{}{
						"Authorization": "token ghp_123456789",
						"Accept":        "application/vnd.github.v3+json",
						"User-Agent":    "GitHub CLI",
					},
					Order: []string{"Authorization", "Accept", "User-Agent"},
				},
				Cookies: map[string]interface{}{
					"logged_in":     true,
					"session_token": "xyz789",
				},
			},
			Response: &sessiondata.ResponseData{
				Body: `{"name": "project", "language": "Go", "stars": 42}`,
			},
		},
		{
			ID: "4",
			Request: &sessiondata.RequestData{
				URL:  "https://httpbin.org/post",
				Body: `{"test": "data", "numbers": [1,2,3]}`,
				Headers: sortedMap.SortedMap{
					Entries: map[string]interface{}{
						"Content-Type":  "application/json",
						"Accept":        "*/*",
						"X-Test-Header": "test-value",
					},
					Order: []string{"Content-Type", "Accept", "X-Test-Header"},
				},
				Cookies: map[string]interface{}{
					"test_cookie": "test123",
				},
			},
			Response: &sessiondata.ResponseData{
				Body: `{"json": {"test": "data"}, "origin": "127.0.0.1"}`,
			},
		},
	}

	for _, session := range sessions {
		store.Store(session)
	}

	return store
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 {
		return true
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func getIDs(sessions []*sessiondata.Session) []string {
	var ids []string
	for _, session := range sessions {
		ids = append(ids, session.ID)
	}
	sort.Strings(ids)
	return ids
}

func TestSearch_BasicFunctionality(t *testing.T) {
	store := setupTestStoreWithData()

	tests := []struct {
		name     string
		options  SearchOptions
		expected []string
		wantErr  bool
	}{
		{
			name:     "Empty options should return error",
			options:  SearchOptions{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Search by URL substring",
			options: SearchOptions{
				URL: "amazon",
			},
			expected: []string{"1"},
			wantErr:  false,
		},
		{
			name: "Search by URL case insensitive",
			options: SearchOptions{
				URL: "AMAZON",
			},
			expected: []string{"1"},
			wantErr:  false,
		},
		{
			name: "Search by URL multiple results",
			options: SearchOptions{
				URL: "https://",
			},
			expected: []string{"1", "2", "3", "4"},
			wantErr:  false,
		},
		{
			name: "Search by URL no results",
			options: SearchOptions{
				URL: "facebook.com",
			},
			expected: []string{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := store.Search(tt.options)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Search() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Search() unexpected error: %v", err)
				return
			}

			var resultIDs []string
			for _, session := range results {
				resultIDs = append(resultIDs, session.ID)
			}

			sort.Strings(resultIDs)
			sort.Strings(tt.expected)

			if !slicesEqual(resultIDs, tt.expected) {
				t.Errorf("Search() = %v, expected %v", resultIDs, tt.expected)
			}
		})
	}
}

func TestSearch_HeadersSearch(t *testing.T) {
	store := setupTestStoreWithData()

	tests := []struct {
		name     string
		options  SearchOptions
		expected []string
	}{
		{
			name: "Search by header key only",
			options: SearchOptions{
				HeadersKey: "Authorization",
			},
			expected: []string{"1", "3"},
		},
		{
			name: "Search by header key case insensitive",
			options: SearchOptions{
				HeadersKey: "authorization",
			},
			expected: []string{"1", "3"},
		},
		{
			name: "Search by header value only",
			options: SearchOptions{
				HeadersVal: "application/json",
			},
			expected: []string{"1", "4"},
		},
		{
			name: "Search by header key and value - both must match independently",
			options: SearchOptions{
				HeadersKey: "Content-Type",
				HeadersVal: "Bearer",
			},
			expected: []string{"1"},
		},
		{
			name: "Search by header key and value - key match, no value match",
			options: SearchOptions{
				HeadersKey: "Accept",
				HeadersVal: "non-existent-value",
			},
			expected: []string{},
		},
		{
			name: "Search by header key and value - value match, no key match",
			options: SearchOptions{
				HeadersKey: "X-Non-Existent",
				HeadersVal: "application/json",
			},
			expected: []string{},
		},
		{
			name: "Search by header key and value - both match in different headers",
			options: SearchOptions{
				HeadersKey: "User-Agent",
				HeadersVal: "application/json",
			},
			expected: []string{"1"},
		},
		{
			name: "Search by header key and value - no matches",
			options: SearchOptions{
				HeadersKey: "X-Non-Existent",
				HeadersVal: "non-existent-value",
			},
			expected: []string{},
		},
		{
			name: "Search by partial header key",
			options: SearchOptions{
				HeadersKey: "User-Agent",
			},
			expected: []string{"1", "3"},
		},
		{
			name: "Search by partial header value",
			options: SearchOptions{
				HeadersVal: "GitHub",
			},
			expected: []string{"3"},
		},
		{
			name: "Search custom header",
			options: SearchOptions{
				HeadersKey: "X-Custom",
			},
			expected: []string{"1"},
		},
		{
			name: "Search by header value substring - Bearer token",
			options: SearchOptions{
				HeadersVal: "Bearer",
			},
			expected: []string{"1"},
		},
		{
			name: "Search by header key substring",
			options: SearchOptions{
				HeadersKey: "Content",
			},
			expected: []string{"1", "2", "4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _ := store.Search(tt.options)

			var resultIDs []string
			for _, session := range results {
				resultIDs = append(resultIDs, session.ID)
			}

			sort.Strings(resultIDs)
			sort.Strings(tt.expected)

			if !slicesEqual(resultIDs, tt.expected) {
				t.Errorf("Search() = %v, expected %v", resultIDs, tt.expected)
			}
		})
	}
}

func TestSearch_CookiesSearch(t *testing.T) {
	store := setupTestStoreWithData()

	tests := []struct {
		name     string
		options  SearchOptions
		expected []string
	}{
		{
			name: "Search by cookie key only",
			options: SearchOptions{
				CookiesKey: "session_id",
			},
			expected: []string{"1"},
		},
		{
			name: "Search by cookie value string",
			options: SearchOptions{
				CookiesVal: "abc123",
			},
			expected: []string{"1"},
		},
		{
			name: "Search by cookie value number",
			options: SearchOptions{
				CookiesVal: 12345,
			},
			expected: []string{"2"},
		},
		{
			name: "Search by cookie value boolean",
			options: SearchOptions{
				CookiesVal: true,
			},
			expected: []string{"3"},
		},
		{
			name: "Search by cookie key and value - both must match independently",
			options: SearchOptions{
				CookiesKey: "user_id",
				CookiesVal: 12345,
			},
			expected: []string{"2"},
		},
		{
			name: "Search by cookie key and value - key match, no value match",
			options: SearchOptions{
				CookiesKey: "session_id",
				CookiesVal: "wrong_value",
			},
			expected: []string{},
		},
		{
			name: "Search by cookie key and value - value match, no key match",
			options: SearchOptions{
				CookiesKey: "non_existent",
				CookiesVal: "abc123",
			},
			expected: []string{},
		},
		{
			name: "Search by cookie key and value - both match in different cookies",
			options: SearchOptions{
				CookiesKey: "session_id",
				CookiesVal: 5,
			},
			expected: []string{"1"},
		},
		{
			name: "Search by partial cookie key",
			options: SearchOptions{
				CookiesKey: "session",
			},
			expected: []string{"1", "3"},
		},
		{
			name: "Search by partial cookie value",
			options: SearchOptions{
				CookiesVal: "theme",
			},
			expected: []string{"1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _ := store.Search(tt.options)

			var resultIDs []string
			for _, session := range results {
				resultIDs = append(resultIDs, session.ID)
			}

			sort.Strings(resultIDs)
			sort.Strings(tt.expected)

			if !slicesEqual(resultIDs, tt.expected) {
				t.Errorf("Search() = %v, expected %v", resultIDs, tt.expected)
			}
		})
	}
}

func TestSearch_BodySearch(t *testing.T) {
	store := setupTestStoreWithData()

	tests := []struct {
		name     string
		options  SearchOptions
		expected []string
	}{
		{
			name: "Search in request body",
			options: SearchOptions{
				Body: "laptop",
			},
			expected: []string{"1"},
		},
		{
			name: "Search in response body",
			options: SearchOptions{
				Body: "documentation",
			},
			expected: []string{"2"},
		},
		{
			name: "Search in both request and response",
			options: SearchOptions{
				Body: "test",
			},
			expected: []string{"4"},
		},
		{
			name: "Case insensitive body search",
			options: SearchOptions{
				Body: "GOLANG",
			},
			expected: []string{"2"},
		},
		{
			name: "JSON content search",
			options: SearchOptions{
				Body: "\"language\":",
			},
			expected: []string{"3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _ := store.Search(tt.options)

			var resultIDs []string
			for _, session := range results {
				resultIDs = append(resultIDs, session.ID)
			}

			sort.Strings(resultIDs)
			sort.Strings(tt.expected)

			if !slicesEqual(resultIDs, tt.expected) {
				t.Errorf("Search() = %v, expected %v", resultIDs, tt.expected)
			}
		})
	}
}

func TestSearch_MultipleCriteria(t *testing.T) {
	store := setupTestStoreWithData()

	tests := []struct {
		name     string
		options  SearchOptions
		expected []string
	}{
		{
			name: "URL and header key - both match",
			options: SearchOptions{
				URL:        "amazon",
				HeadersKey: "Content-Type",
			},
			expected: []string{"1"},
		},
		{
			name: "URL and header key - URL matches, header doesn't",
			options: SearchOptions{
				URL:        "amazon",
				HeadersKey: "X-Missing-Header",
			},
			expected: []string{},
		},
		{
			name: "URL, header and cookie - all match",
			options: SearchOptions{
				URL:        "google",
				HeadersKey: "Accept",
				CookiesKey: "lang",
			},
			expected: []string{"2"},
		},
		{
			name: "URL, header and body - all match",
			options: SearchOptions{
				URL:        "github",
				HeadersKey: "Authorization",
				Body:       "project",
			},
			expected: []string{"3"},
		},
		{
			name: "All criteria - match",
			options: SearchOptions{
				URL:        "httpbin",
				HeadersKey: "Content-Type",
				HeadersVal: "application/json",
				CookiesKey: "test_cookie",
				Body:       "json",
			},
			expected: []string{"4"},
		},
		{
			name: "All criteria - header key/value both match",
			options: SearchOptions{
				URL:        "amazon",
				HeadersKey: "Authorization",
				HeadersVal: "application/json",
				CookiesKey: "session_id",
				Body:       "laptop",
			},
			expected: []string{"1"},
		},
		{
			name: "Headers key and value independent - value missing",
			options: SearchOptions{
				HeadersKey: "Content-Type",
				HeadersVal: "non-existent-value",
			},
			expected: []string{},
		},
		{
			name: "Headers key and value independent - key missing",
			options: SearchOptions{
				HeadersKey: "X-Missing",
				HeadersVal: "application/json",
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _ := store.Search(tt.options)

			var resultIDs []string
			for _, session := range results {
				resultIDs = append(resultIDs, session.ID)
			}

			sort.Strings(resultIDs)
			sort.Strings(tt.expected)

			if !slicesEqual(resultIDs, tt.expected) {
				t.Errorf("Search() = %v, expected %v", resultIDs, tt.expected)
			}
		})
	}
}

func BenchmarkSearch_URLOnly(b *testing.B) {
	store := setupTestStoreWithData()
	opt := SearchOptions{URL: "amazon"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Search(opt)
	}
}

func BenchmarkSearch_HeadersOnly(b *testing.B) {
	store := setupTestStoreWithData()
	opt := SearchOptions{HeadersKey: "Content-Type"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Search(opt)
	}
}

func BenchmarkSearch_MultipleCriteria(b *testing.B) {
	store := setupTestStoreWithData()
	opt := SearchOptions{
		URL:        "github",
		HeadersKey: "Authorization",
		Body:       "project",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Search(opt)
	}
}
