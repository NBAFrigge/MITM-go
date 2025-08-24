package bodyParser

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
)

const (
	ContentTypeJSON    = "application/json"
	CompressionGzip    = "gzip"
	CompressionZstd    = "zstd"
	CompressionDeflate = "deflate"
)

type BodyParserOptions struct {
	ContentType string
	Compression string
}

func NewBodyParserOptions() BodyParserOptions {
	return BodyParserOptions{}
}

func (opt *BodyParserOptions) PopulateFromHeaders(headers map[string][]string) {
	if ct, ok := headers["Content-Type"]; ok && len(ct) > 0 {
		opt.ContentType = ct[0]
	}
	if comp, ok := headers["Content-Encoding"]; ok && len(comp) > 0 {
		opt.Compression = comp[0]
	}
}

func Parse(body string, options BodyParserOptions) (string, error) {
	decompressedBody, err := decompress(body, options.Compression)
	if err != nil {
		return "", err
	}

	if strings.Contains(options.ContentType, ContentTypeJSON) {
		return formatJSON(decompressedBody)
	}

	return decompressedBody, nil
}

func decompress(body, compression string) (string, error) {
	switch compression {
	case CompressionGzip:
		reader, err := gzip.NewReader(bytes.NewReader([]byte(body)))
		if err != nil {
			return "", fmt.Errorf("gzip decompression error: %w", err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("reading decompressed data: %w", err)
		}
		return string(decompressed), nil
	case CompressionZstd:
		reader, err := zstd.NewReader(bytes.NewReader([]byte(body)))
		if err != nil {
			return "", fmt.Errorf("zstd decompression error: %w", err)
		}
		defer reader.Close()
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("reading decompressed data: %w", err)
		}
		return string(decompressed), nil
	case CompressionDeflate:
		reader := flate.NewReader(bytes.NewReader([]byte(body)))
		defer reader.Close()
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return "", fmt.Errorf("deflate decompression error: %w", err)
		}
		return string(decompressed), nil
	case "":
		return body, nil
	default:
		return "", fmt.Errorf("unsupported compression method: %s", compression)
	}
}

func formatJSON(body string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return "", err
	}
	parsedBody, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(parsedBody), nil
}
