# MITM-go

Terminal-based HTTP/HTTPS debugging proxy built in Go. Intercepts, inspects, and replays traffic in real-time with TLS fingerprint extraction.

## Features

- **HTTPS Interception** — MITM proxy with dynamic certificate generation
- **HTTP/2** — Full support including HPACK decoding and frame analysis
- **TLS Fingerprinting** — Extracts JA3 hash, cipher suites, extensions, curves, signature algorithms from the original ClientHello
- **WebSocket** — Real-time interception and visualization of messages
- **Header Order Preservation** — Custom parser that maintains original header ordering
- **Body Handling** — Automatic decompression (Gzip, Deflate, Zstd) and JSON formatting
- **Request Replay** — Re-send captured requests through the proxy
- **cURL Export** — Copy any session as a cURL command
- **Regex Filtering** — Filter sessions by URL pattern

## Requirements

- Go 1.24+

## Install

```bash
go mod tidy
go build -o mitm-go .
```

## Usage

```bash
./mitm-go
./mitm-go -port 9090
```

Configure your client to use `http://127.0.0.1:8080` as proxy. Install `certs/httpCA.crt` as a trusted CA to intercept HTTPS.

## Keybindings

| Key      | Action                            |
| -------- | --------------------------------- |
| `Ctrl+S` | Start/stop proxy                  |
| `Enter`  | Select session / toggle details   |
| `Tab`    | Switch panel focus                |
| `←→`     | Switch tab (Request/Response/TLS) |
| `↑↓`     | Navigate / scroll                 |
| `/`      | Search (regex filter by URL)      |
| `r`      | Replay selected request           |
| `c`      | Copy as cURL                      |
| `Ctrl+D` | Clear all sessions                |
| `Ctrl+R` | Refresh sessions                  |
| `F1`     | Help                              |
| `F2`     | Toggle verbose logging            |
| `Esc`    | Close details / clear filter      |
| `q`      | Quit                              |

## License

MIT
