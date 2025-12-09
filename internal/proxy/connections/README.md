# Connections Package

This package contains specialized implementations of `net.Conn` and `net.Listener` to facilitate traffic interception, buffering, and analysis without disrupting the underlying TCP stream.

## Components

### HTTP2FrameWrapper
A robust wrapper around a `net.Conn` that decodes HTTP/2 frames on the fly.
* **HPACK Decoding**: Uses `golang.org/x/net/http2/hpack` to decode headers from `HEADERS` and `CONTINUATION` frames.
* **Frame Analysis**: Identifies frame types (DATA, HEADERS, SETTINGS, etc.) and stream IDs.
* **Callbacks**: Triggers callbacks when headers are fully decoded, allowing the MITM handler to capture HTTP/2 metadata that is otherwise opaque in the encrypted stream.

### Wrapper Types
* **BufferedConn**: Allows peeking at the initial bytes of a connection (e.g., to determine if it is TLS) before handing the connection off to the TLS handshake library.
* **ReplayConn**: Injects previously read bytes back into the read stream. Used to "replay" the Client Hello message during the TLS handshake process after analysis.
* **CapturingConn**: Copies all read data into a `bytes.Buffer`. Used to capture the raw HTTP/1.1 request wire format.
* **SingleConnListener**: An adapter that turns a single `net.Conn` into a `net.Listener`. This allows the `http.Server` or `http2.Server` to serve a single hijacked connection.
