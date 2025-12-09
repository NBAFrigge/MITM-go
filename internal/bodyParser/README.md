# Body Parser Package

Handles the transformation of request/response bodies into a readable format.

## Public Methods

- **`NewBodyParserOptions() BodyParserOptions`**
  Creates an empty options structure.

- **`PopulateFromHeaders(headers map[string][]string)`**
  Method of the `BodyParserOptions` structure. Analyzes `Content-Type` and `Content-Encoding` headers to automatically configure the parser (e.g., detecting if content is Gzipped or JSON).

- **`Parse(body string, options BodyParserOptions) (string, error)`**
  Main transformation function.
  1. **Decompression**: Supports `gzip`, `zstd` (via `klauspost/compress`), and `deflate`.
  2. **Formatting**: If Content-Type is `application/json`, performs Unmarshal/Marshal with indentation to make JSON readable (pretty-print).
  3. Returns the processed string.
