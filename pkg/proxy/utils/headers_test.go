package utils

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestCleanHeader_PassesNormalHeaders(t *testing.T) {
	src := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer token"},
		"Accept":        {"*/*"},
	}
	dst := http.Header{}
	CleanHeader(dst, src)

	if dst.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type not preserved: %q", dst.Get("Content-Type"))
	}
	if dst.Get("Authorization") != "Bearer token" {
		t.Errorf("Authorization not preserved: %q", dst.Get("Authorization"))
	}
}

func TestCleanHeader_FiltersHopByHop(t *testing.T) {
	src := http.Header{
		"Connection":        {"keep-alive"},
		"Proxy-Connection":  {"keep-alive"},
		"Upgrade":           {"websocket"},
		"Transfer-Encoding": {"chunked"},
		"Content-Type":      {"text/plain"},
	}
	dst := http.Header{}
	CleanHeader(dst, src)

	for _, hop := range []string{"Connection", "Proxy-Connection", "Upgrade", "Transfer-Encoding"} {
		if _, ok := dst[hop]; ok {
			t.Errorf("hop-by-hop header %q should be filtered", hop)
		}
	}
	if dst.Get("Content-Type") != "text/plain" {
		t.Error("Content-Type should pass through")
	}
}

func TestCleanHeader_FiltersHopByHopCaseInsensitive(t *testing.T) {
	src := http.Header{
		"connection":         {"close"},
		"UPGRADE":            {"h2c"},
		"transfer-encoding":  {"gzip"},
	}
	dst := http.Header{}
	CleanHeader(dst, src)

	for _, k := range []string{"connection", "UPGRADE", "transfer-encoding"} {
		if _, ok := dst[http.CanonicalHeaderKey(k)]; ok {
			t.Errorf("header %q should be filtered", k)
		}
	}
}

func TestCleanHeader_ContentLength_Present(t *testing.T) {
	src := http.Header{"Content-Length": {"100"}}
	dst := http.Header{}
	CleanHeader(dst, src)

	if dst.Get("Content-Length") != "100" {
		t.Errorf("Content-Length = %q, want 100", dst.Get("Content-Length"))
	}
}

func TestCleanHeader_ContentLength_EmptySliceDeletes(t *testing.T) {
	src := http.Header{"Content-Length": {}}
	dst := http.Header{"Content-Length": {"existing"}}
	CleanHeader(dst, src)

	if dst.Get("Content-Length") != "" {
		t.Errorf("Content-Length should be deleted for empty slice, got %q", dst.Get("Content-Length"))
	}
}

func TestWriteFilteredHeaders_WritesNormalHeaders(t *testing.T) {
	headers := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer token"},
	}

	var buf bytes.Buffer
	if err := WriteFilteredHeaders(&buf, headers); err != nil {
		t.Fatalf("WriteFilteredHeaders failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Content-Type: application/json\r\n") {
		t.Errorf("Content-Type not in output: %q", output)
	}
	if !strings.Contains(output, "Authorization: Bearer token\r\n") {
		t.Errorf("Authorization not in output: %q", output)
	}
}

func TestWriteFilteredHeaders_FiltersHopByHop(t *testing.T) {
	headers := http.Header{
		"Content-Type":      {"text/plain"},
		"Connection":        {"keep-alive"},
		"Transfer-Encoding": {"chunked"},
		"Upgrade":           {"websocket"},
	}

	var buf bytes.Buffer
	WriteFilteredHeaders(&buf, headers)
	output := buf.String()

	for _, hop := range []string{"Connection:", "Transfer-Encoding:", "Upgrade:"} {
		if strings.Contains(output, hop) {
			t.Errorf("hop-by-hop header %q should not be in output", hop)
		}
	}
	if !strings.Contains(output, "Content-Type:") {
		t.Error("Content-Type should be in output")
	}
}

func TestWriteFilteredHeaders_MultipleValues(t *testing.T) {
	headers := http.Header{
		"X-Custom": {"value1", "value2"},
	}

	var buf bytes.Buffer
	WriteFilteredHeaders(&buf, headers)
	output := buf.String()

	if !strings.Contains(output, "X-Custom: value1") {
		t.Error("first value should be present")
	}
	if !strings.Contains(output, "X-Custom: value2") {
		t.Error("second value should be present")
	}
}
