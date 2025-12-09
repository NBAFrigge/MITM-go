# Client Hello Package

This package is essential for "TLS Fingerprinting" techniques. It analyzes raw bytes of the initial TLS message (`ClientHello`) to extract parameters identifying the client (browser, bot, script).

## Public Methods

### `ClientHelloParser`

- **`ParseClientHello(rawClientHello []byte) (*tls.Config, error)`**
  Manually parses the byte slice of the TLS Handshake packet (0x16). Extracts:
  - TLS Version (Min/Max).
  - Supported Cipher Suites (in order).
  - TLS Extensions, with specific focus on: SNI (Server Name Indication), Supported Elliptic Curves, Signature Algorithms, ALPN, and Supported Versions.
  - Handles and filters GREASE values (random values reserved to test extensibility).
    Returns a `tls.Config` that mimics the original client configuration.

- **`GenerateClientHelloHash(data []byte) string`**
  Generates a hash (FNV-64a) representative of the ClientHello. To optimize performance on large payloads, it hashes a central subsection of the bytes and the total length.

### `ClientHelloCache`

- **`NewClientHelloCache() *ClientHelloCache`**
  Initializes the in-memory store for TLS configurations.

- **`Get(key []byte) (*tls.Config, bool)`** / **`Set(key []byte, config *tls.Config)`**
  Thread-safe methods to save and retrieve TLS configurations based on the raw ClientHello hash.
