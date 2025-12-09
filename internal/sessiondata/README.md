# SessionData Package

Defines data models and business logic associated with a single transaction (Session).

## Public Methods

### `NewSessionData`

Factory constructor that normalizes an `http.Request` and its raw bytes into a `Session` structure. Automatically determines whether to create a standard HTTP session or a WebSocket one based on Upgrade headers.

### `Session` Struct Methods

- **`CompareRequest(other *Session) bool`**
  Quickly compares two sessions to verify basic equality (Method, URL, Content-Type, Body, and Headers).

- **`RequestDifferences(other *Session) *RequestDifference`**
  Performs a deep analysis (Deep Diff) between two sessions. Returns a detailed structure highlighting:
  - Changes in scalar fields (Method, URL).
  - Added, removed, or modified headers.
  - Added, removed, or modified cookies.
  - Differences in the Body.

- **`ToCurl() string`**
  Generates a valid cURL command string to reproduce the request.
  - Uses `sortedMap` to ensure the header order in the cURL command matches exactly that of the captured request.
  - Excludes automatic headers (like `Content-Length` or `Host`) to avoid conflicts.

- **`Replay(port int) error`**
  Re-executes the captured request through the proxy itself.
  - If HTTPS: uses the `tls-client` library (bogdanfinn) to simulate a realistic TLS fingerprint.
  - If HTTP: uses the standard Go client.
  - Preserves original body, headers, and cookies.
