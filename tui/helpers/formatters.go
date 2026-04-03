package helpers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fµs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1000000)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}

func FormatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	now := time.Now()
	if t.Format("2006-01-02") == now.Format("2006-01-02") {
		return t.Format("15:04:05")
	}

	return t.Format("02/01 15:04")
}

func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

func FormatHTTPStatus(code int) string {
	if code == 0 {
		return "-"
	}

	text := http.StatusText(code)
	if text == "" {
		return fmt.Sprintf("%d", code)
	}

	return fmt.Sprintf("%d %s", code, text)
}

func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}

	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}

	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}

func FormatHeaders(headers http.Header) string {
	if len(headers) == 0 {
		return "No headers"
	}

	var parts []string
	for name, values := range headers {
		for _, value := range values {
			parts = append(parts, fmt.Sprintf("%s: %s", name, value))
		}
	}

	return strings.Join(parts, "\n")
}

func FormatURL(url string, maxLen int) string {
	if url == "" {
		return "-"
	}

	display := url
	if strings.HasPrefix(url, "https://") {
		display = strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		display = strings.TrimPrefix(url, "http://")
	}
	return TruncateString(display, maxLen)
}

func FormatMethod(method string) string {
	if method == "" {
		return "    -"
	}

	switch len(method) {
	case 3:
		return method + "  "
	case 4:
		return method + " "
	case 5:
		return method
	case 6:
		return method
	default:
		return TruncateString(method, 6)
	}
}

func FormatContentType(contentType string) string {
	if contentType == "" {
		return "-"
	}

	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}

	switch contentType {
	case "application/json":
		return "JSON"
	case "application/xml", "text/xml":
		return "XML"
	case "text/html":
		return "HTML"
	case "text/plain":
		return "Text"
	case "application/octet-stream":
		return "Binary"
	case "multipart/form-data":
		return "Form"
	case "application/x-www-form-urlencoded":
		return "URLForm"
	default:
		return TruncateString(contentType, 15)
	}
}

func FormatBodyPreview(body []byte, maxLen int) string {
	if len(body) == 0 {
		return "Empty"
	}

	bodyStr := string(body)
	if !isPrintable(bodyStr) {
		return fmt.Sprintf("Binary (%s)", FormatBytes(int64(len(body))))
	}

	cleaned := strings.ReplaceAll(bodyStr, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	cleaned = strings.TrimSpace(cleaned)

	return TruncateString(cleaned, maxLen)
}

func isPrintable(s string) bool {
	if len(s) == 0 {
		return true
	}

	printable := 0
	total := 0

	for _, r := range s {
		total++
		if r >= 32 && r < 127 || r == '\n' || r == '\r' || r == '\t' {
			printable++
		}
	}

	return float64(printable)/float64(total) >= 0.8
}
