# Utils Package

This package provides helper functions for low-level HTTP operations within the proxy.

## Key Functions

* **HandleProxyError**: Centralized error reporting. Updates the Session state with the error and sends an appropriate HTTP error response (502 Bad Gateway, etc.) to the client.
* **CleanHeader**: Removes hop-by-hop headers (e.g., `Connection`, `Proxy-Connection`, `Transfer-Encoding`) defined in RFC 2616 before forwarding requests or responses.
* **ProcessAndStoreHTTPSession**: Orchestrates the full lifecycle of an HTTP request: logging, forwarding, response reading, body parsing, and storage in the database.
* **WriteHTTPResponse**: Writes a `http.Response` object to a raw `io.Writer` (used in hijacked connections).
