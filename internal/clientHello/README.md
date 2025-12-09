# Client Hello Package

The `clientHello` package is responsible for analyzing raw TLS Client Hello messages. This analysis is critical for "TLS Signature Replication," allowing the proxy to impersonate the downstream client's TLS fingerprint when connecting to the upstream server.

## Components

### ClientHelloParser
Implements a low-level byte parser for the TLS Handshake protocol (Record Type 0x16, Handshake Type 0x01). It extracts:
* **TLS Version**: Min and Max supported versions.
* **Cipher Suites**: The ordered list of supported cipher suites.
* **Extensions**: Specifically parses `SNI`, `SupportedCurves`, `SignatureAlgorithms`, `ALPN`, and `SupportedVersions`.

### ClientHelloCache
Provides a concurrent-safe caching mechanism (`sync.RWMutex`) for `tls.Config` objects derived from Client Hello messages.
* **Key**: A hash of the Client Hello bytes (using FNV-64a). To optimize for large payloads, a subset of the bytes is hashed.
* **Value**: A pre-configured `crypto/tls.Config` object ready for use by the upstream client.

## Technical Details
The parser manually decodes the variable-length vectors defined in RFC 5246 and RFC 8446. It handles GREASE (Generate Random Extensions And Sustain Extensibility) values by filtering them out during the version translation to standard Go `crypto/tls` constants.
