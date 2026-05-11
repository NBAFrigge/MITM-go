package headerParser

import (
	"testing"
)

func TestParseHeadersFromRaw_Basic(t *testing.T) {
	raw := []byte("GET /path HTTP/1.1\r\nHost: example.com\r\nContent-Type: application/json\r\nAuthorization: Bearer token\r\n\r\n")

	headers, err := ParseHeadersFromRaw(raw)
	if err != nil {
		t.Fatalf("ParseHeadersFromRaw failed: %v", err)
	}

	tests := []struct{ key, want string }{
		{"Host", "example.com"},
		{"Content-Type", "application/json"},
		{"Authorization", "Bearer token"},
	}
	for _, tt := range tests {
		val, ok := headers.Get(tt.key)
		if !ok {
			t.Errorf("header %q not found", tt.key)
			continue
		}
		if val != tt.want {
			t.Errorf("header %q = %v, want %q", tt.key, val, tt.want)
		}
	}
}

func TestParseHeadersFromRaw_EmptyHeaders(t *testing.T) {
	raw := []byte("GET / HTTP/1.1\r\n\r\n")

	headers, err := ParseHeadersFromRaw(raw)
	if err != nil {
		t.Fatalf("ParseHeadersFromRaw failed: %v", err)
	}
	if headers.Len() != 0 {
		t.Errorf("expected 0 headers, got %d", headers.Len())
	}
}

func TestParseHeadersFromRaw_ColonInValue(t *testing.T) {
	raw := []byte("GET / HTTP/1.1\r\nX-Custom: value:with:colons\r\n\r\n")

	headers, err := ParseHeadersFromRaw(raw)
	if err != nil {
		t.Fatalf("ParseHeadersFromRaw failed: %v", err)
	}

	val, ok := headers.Get("X-Custom")
	if !ok {
		t.Fatal("X-Custom header not found")
	}
	if val != "value:with:colons" {
		t.Errorf("X-Custom = %v, want value:with:colons", val)
	}
}

func TestParseHeadersFromRaw_PreservesOrder(t *testing.T) {
	raw := []byte("POST /api HTTP/1.1\r\nA: 1\r\nB: 2\r\nC: 3\r\n\r\n")

	headers, err := ParseHeadersFromRaw(raw)
	if err != nil {
		t.Fatalf("ParseHeadersFromRaw failed: %v", err)
	}

	order := headers.Keys()
	expected := []string{"A", "B", "C"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d headers, got %d", len(expected), len(order))
	}
	for i, k := range expected {
		if order[i] != k {
			t.Errorf("Order[%d] = %q, want %q", i, order[i], k)
		}
	}
}

func TestParseHeadersFromRaw_TrimsWhitespace(t *testing.T) {
	raw := []byte("GET / HTTP/1.1\r\nX-Header:   trimmed value   \r\n\r\n")

	headers, err := ParseHeadersFromRaw(raw)
	if err != nil {
		t.Fatalf("ParseHeadersFromRaw failed: %v", err)
	}

	val, ok := headers.Get("X-Header")
	if !ok {
		t.Fatal("X-Header not found")
	}
	if val != "trimmed value" {
		t.Errorf("X-Header = %q, want %q", val, "trimmed value")
	}
}

func BenchmarkParseHeadersFromRaw(b *testing.B) {
	raw := []byte("GET /api/v1/users HTTP/1.1\r\nHost: api.example.com\r\nContent-Type: application/json\r\nAuthorization: Bearer token123\r\nAccept: application/json\r\nUser-Agent: TestClient/1.0\r\n\r\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseHeadersFromRaw(raw)
	}
}
