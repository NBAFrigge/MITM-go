# Handlers Package

This package implements the core logic for processing intercepted traffic based on the protocol type.

## Handlers

### Manager
Orchestrates the instantiation of specific handlers and holds references to global configuration and caches.

### HTTPHandler
Handles standard, clear-text HTTP requests acting as a forward proxy.
* Parses the request body.
* Creates a `Session` object.
* Forwards the request using the configured HTTP client.

### MITMHandler
Handles `CONNECT` requests to establish HTTPS tunnels.
1.  **Hijacking**: Takes over the raw TCP connection from the client.
2.  **Client Hello Analysis**: Peeks at the initial bytes to parse the Client Hello and determine TLS fingerprinting parameters.
3.  **Certificate Generation**: Uses the `certs` package to generate a certificate for the requested SNI.
4.  **TLS Handshake**: Completes the handshake with the client.
5.  **Protocol Negotiation**: Checks `ALPN` (Application-Layer Protocol Negotiation) to determine if the traffic is HTTP/1.1 or HTTP/2.
6.  **HTTP/1.1 Handling**: Uses `connections.CapturingConn` to read the request, parses it, and forwards it.
7.  **HTTP/2 Handling**: Initializes a `golang.org/x/net/http2.Server` over the connection. Uses `connections.HTTP2FrameWrapper` to capture headers and data frames associated with specific streams.

### WebSocketHandler
Handles the WebSocket Upgrade handshake and subsequent frame forwarding.
* **Interception**: Intercepts the HTTP Upgrade request.
* **Bidirectional Forwarding**: Launches two goroutines to copy frames between the client and the backend.
* **Frame Parsing**: Decodes WebSocket frames (Text, Binary, Ping, Pong, Close) to store message content and statistics in the `Session` object.
* **Masking**: Handles unmasking of client-to-server frames and masking of server-to-client frames if required by the protocol.
