# Proxy Package

The `proxy` package orchestrates traffic interception. It implements the `http.Handler` interface and routes traffic based on the method (standard or CONNECT) and the negotiated protocol.

## Public Methods and Structures

### `Proxy`

The main structure encapsulating configuration, session store, and handlers.

- **`NewProxy(store interfaces.SessionStore, logger interfaces.Logger, caCache *certs.CertCache) *Proxy`**
  Initializes a new proxy instance. It configures the HTTP transport client (with TLS 1.2+ support and automatic compression disabled to handle it manually) and instantiates the handler `Manager`.

- **`ServeHTTP(w http.ResponseWriter, r *http.Request)`**
  Entry point for every request.
  - If the method is `CONNECT`: delegates management to `MITMHandler` to establish the HTTPS tunnel.
  - Otherwise: delegates to `HTTPHandler` for cleartext proxying.

---

## Subpackage: Handlers (`internal/proxy/handlers`)

### `Manager`

Coordinates the different types of handlers (HTTP, MITM, WebSocket).

- **`HandleHTTP`**, **`HandleMITM`**: Wrappers to invoke the respective specialized handlers.

### `HTTPHandler`

Handles cleartext HTTP traffic.

- **`Handle`**: Reads the request body, creates a `Session` object, forwards the request upstream, and returns the response to the client, saving data to the store.

### `MITMHandler`

Handles intercepted HTTPS traffic.

- **`Handle`**:
  1. Hijacks the TCP connection.
  2. Sends `200 Connection Established`.
  3. Reads the first bytes to intercept the TLS `ClientHello`.
  4. Uses `clientHelloParser` to extract client cryptographic preferences (useful for fingerprinting).
  5. Generates a dynamic certificate for the requested host using `certs.CertCache`.
  6. Performs the TLS handshake with the client.
  7. Based on the negotiated ALPN (`h2` or `http/1.1`), starts specific handling (HTTP/2 server or HTTP/1.1 read loop).

### `WebSocketHandler`

Handles upgrade and WebSocket tunneling.

- **`Handle`**:
  1. Intercepts the upgrade handshake.
  2. Connects to the backend (using TLS if necessary).
  3. If the handshake succeeds, starts two goroutines for bidirectional frame piping.
  4. Decodes each passing frame (opcode, payload, masking) to update session statistics and logs in real-time.

---

## Subpackage: Connections (`internal/proxy/connections`)

Provides wrappers around `net.Conn` for specific purposes.

- **`HTTP2FrameWrapper`**: Decodes HTTP/2 frames (HEADERS, DATA, etc.) using HPACK as they pass through the connection, allowing request reconstruction within the encrypted tunnel.
- **`BufferedConn` / `ReplayConn`**: Allow "peeking" at data (e.g., ClientHello) and reinserting it into the read stream so the standard TLS library can process it.
- **`CapturingConn`**: Captures all raw read traffic to allow manual header parsing.
