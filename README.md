# MITM-go - HTTP Debugger & Proxy

MITM-go is an advanced tool for intercepting, inspecting, and debugging HTTP(S) traffic in real-time. Built with a Go backend and a Vue.js frontend (via Wails), it allows for detailed analysis of communications between client and server, supporting modern protocols like HTTP/2 and WebSockets.

## Key Features

- **HTTP/HTTPS Interception**: Acts as a Man-in-the-Middle proxy, decrypting SSL/TLS traffic on-the-fly via dynamic certificate generation.
- **HTTP/2 Support**: Full management of HTTP/2 multiplexing, including HPACK decoding and frame analysis.
- **TLS Fingerprint Replication**: Analysis of the original `ClientHello` to replicate the TLS signature (JA3) in upstream connections, evading detection by anti-bot systems.
- **WebSocket Support**: Real-time interception, decoding, and visualization of WebSocket messages (text and binary).
- **Header Ordering**: Preservation of the original order of HTTP headers (critical for fingerprinting analysis) via a custom parser.
- **Body Analysis**: Automatic decompression (Gzip, Deflate, Zstd) and JSON formatting.
- **Session Management**:
  - Advanced search (URL, Header, Cookie, Body).
  - Differential comparison (Diff) between two requests.
  - Request replay.
  - Export to cURL format.

## Build Requirements

- **Go**: 1.24+
- **Node.js & NPM**: For frontend building.
- **Wails**: CLI installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

## Installation Instructions

```bash
# Download Go dependencies
go mod tidy

# Install and build the frontend
cd frontend
npm install
npm run build
cd ..

# Build the desktop application
wails build
```

## License

Distributed under the MIT License.
