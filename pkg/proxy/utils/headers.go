package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CleanHeader copies headers from src to dst, excluding hop-by-hop headers
func CleanHeader(dst, src http.Header) {
	for k, v := range src {
		switch strings.ToLower(k) {
		case "connection", "proxy-connection", "upgrade", "transfer-encoding":
			continue
		case "content-length":
			if len(v) > 0 {
				dst.Set(k, v[0])
			} else {
				dst.Del(k)
			}
		default:
			dst[k] = v
		}
	}
}

// WriteFilteredHeaders writes headers to the connection, excluding hop-by-hop headers
func WriteFilteredHeaders(conn io.Writer, headers http.Header) error {
	for k, v := range headers {
		switch strings.ToLower(k) {
		case "connection", "proxy-connection", "upgrade", "transfer-encoding":
			continue
		default:
			for _, vv := range v {
				_, err := fmt.Fprintf(conn, "%s: %s\r\n", k, vv)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
