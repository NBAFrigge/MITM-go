package session

import "testing"

func TestMatchString_Substring(t *testing.T) {
	tests := []struct {
		str, pattern string
		want         bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "xyz", false},
		{"hello world", "", true},
		{"", "pattern", false},
		{"", "", true},
	}
	for _, tt := range tests {
		got := matchString(tt.str, tt.pattern)
		if got != tt.want {
			t.Errorf("matchString(%q, %q) = %v, want %v", tt.str, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchString_CaseInsensitive(t *testing.T) {
	tests := []struct {
		str, pattern string
		want         bool
	}{
		{"Hello World", "WORLD", true},
		{"HELLO", "hello", true},
		{"Bearer Token123", "bearer", true},
	}
	for _, tt := range tests {
		got := matchString(tt.str, tt.pattern)
		if got != tt.want {
			t.Errorf("matchString(%q, %q) = %v, want %v", tt.str, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchString_Regex(t *testing.T) {
	tests := []struct {
		str, pattern string
		want         bool
	}{
		{"hello world", "/hel+o/", true},
		{"hello world", "/^world/", false},
		{"hello world", "/world$/", true},
		{"user123", "/user\\d+/", true},
		{"abc", "/^abc$/", true},
		{"abcd", "/^abc$/", false},
		{"test", "/foo|test/", true},
	}
	for _, tt := range tests {
		got := matchString(tt.str, tt.pattern)
		if got != tt.want {
			t.Errorf("matchString(%q, %q) = %v, want %v", tt.str, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchString_InvalidRegexFallsBackToContains(t *testing.T) {
	got := matchString("hello world", "/[invalid/")
	if got {
		t.Error("expected no match for invalid regex with no literal match")
	}

	got = matchString("contains /[invalid/ literally", "/[invalid/")
	if !got {
		t.Error("invalid regex should fall back to literal contains of pattern string")
	}
}

func TestIsRegex(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"/regex/", true},
		{"/abc/", true},
		{"regex", false},
		{"/noclose", false},
		{"noopen/", false},
		{"//", false},
		{"/x/", true},
		{"plain text", false},
	}
	for _, tt := range tests {
		got := isRegex(tt.query)
		if got != tt.want {
			t.Errorf("isRegex(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}
