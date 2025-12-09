# Body Parser Package

The `bodyParser` package provides functionality to process raw HTTP body content based on the `Content-Type` and `Content-Encoding` headers.

## Functionality

The primary entry point is the `Parse` function, which accepts the raw body string and a `BodyParserOptions` struct.

### Decompression
The package supports automatic decompression of the following content encodings:
* **gzip**: Uses `compress/gzip`.
* **zstd**: Uses `github.com/klauspost/compress/zstd`.
* **deflate**: Uses `compress/flate`.

### Formatting
* **JSON**: If the content type indicates JSON (`application/json`), the payload is unmarshaled and indented using `encoding/json` to provide a human-readable string.

## Structures

* **BodyParserOptions**: Configures the parsing behavior, typically populated directly from HTTP headers via `PopulateFromHeaders`.
