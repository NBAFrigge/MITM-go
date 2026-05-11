package bodyParser

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

func TestParse_PlainText(t *testing.T) {
	result, err := Parse("hello world", BodyParserOptions{})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if result != "hello world" {
		t.Errorf("Parse() = %q, want %q", result, "hello world")
	}
}

func TestParse_EmptyBody(t *testing.T) {
	result, err := Parse("", BodyParserOptions{})
	if err != nil {
		t.Fatalf("Parse empty body failed: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestParse_JSON_FormatsPretty(t *testing.T) {
	input := `{"name":"test","value":42}`
	result, err := Parse(input, BodyParserOptions{ContentType: "application/json"})
	if err != nil {
		t.Fatalf("Parse JSON failed: %v", err)
	}
	if !strings.Contains(result, "\n") {
		t.Errorf("expected indented JSON, got %q", result)
	}
	if !strings.Contains(result, `"name"`) {
		t.Errorf("expected key name in output, got %q", result)
	}
}

func TestParse_JSON_InvalidReturnsError(t *testing.T) {
	_, err := Parse("not json", BodyParserOptions{ContentType: "application/json"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParse_JSON_ContentTypeSubstring(t *testing.T) {
	input := `{"ok":true}`
	result, err := Parse(input, BodyParserOptions{ContentType: "application/json; charset=utf-8"})
	if err != nil {
		t.Fatalf("Parse with content-type params failed: %v", err)
	}
	if !strings.Contains(result, `"ok"`) {
		t.Errorf("expected formatted JSON, got %q", result)
	}
}

func TestParse_Gzip(t *testing.T) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write([]byte("compressed content"))
	w.Close()

	result, err := Parse(buf.String(), BodyParserOptions{Compression: CompressionGzip})
	if err != nil {
		t.Fatalf("Parse gzip failed: %v", err)
	}
	if result != "compressed content" {
		t.Errorf("Parse gzip = %q, want %q", result, "compressed content")
	}
}

func TestParse_Gzip_InvalidData(t *testing.T) {
	_, err := Parse("not gzip data", BodyParserOptions{Compression: CompressionGzip})
	if err == nil {
		t.Error("expected error for invalid gzip data")
	}
}

func TestParse_UnsupportedCompression(t *testing.T) {
	_, err := Parse("data", BodyParserOptions{Compression: "lz4"})
	if err == nil {
		t.Error("expected error for unsupported compression")
	}
}

func TestPopulateFromHeaders(t *testing.T) {
	opt := NewBodyParserOptions()
	headers := map[string][]string{
		"Content-Type":     {"application/json; charset=utf-8"},
		"Content-Encoding": {"gzip"},
	}
	opt.PopulateFromHeaders(headers)

	if !strings.Contains(opt.ContentType, "application/json") {
		t.Errorf("ContentType = %q, want to contain application/json", opt.ContentType)
	}
	if opt.Compression != "gzip" {
		t.Errorf("Compression = %q, want gzip", opt.Compression)
	}
}

func TestPopulateFromHeaders_Empty(t *testing.T) {
	opt := NewBodyParserOptions()
	opt.PopulateFromHeaders(map[string][]string{})
	if opt.ContentType != "" || opt.Compression != "" {
		t.Error("expected empty options for empty headers")
	}
}

func TestPopulateFromHeaders_NoEncoding(t *testing.T) {
	opt := NewBodyParserOptions()
	opt.PopulateFromHeaders(map[string][]string{
		"Content-Type": {"text/plain"},
	})
	if opt.Compression != "" {
		t.Errorf("expected no compression, got %q", opt.Compression)
	}
}

func BenchmarkParse_JSON(b *testing.B) {
	input := `{"users":[{"id":1,"name":"alice"},{"id":2,"name":"bob"}]}`
	opt := BodyParserOptions{ContentType: "application/json"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parse(input, opt)
	}
}
