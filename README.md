# MITM-go

HTTP Debugger is a tool for inspecting and debugging HTTP(S) requests and responses.
It allows developers to monitor, analyze, and troubleshoot HTTP traffic in real-time.

---

## Features
- Capture and display HTTP/HTTPS requests and responses
- Capture messages from websockets 
- Display real request's headers order extracting them from raw data
- Support for HTTP/2
- Client TLS Signature Replication
- View request and response headers, body, and status codes
- search traffic by URL, headers key or value, cookies
- Confront 2 different HTTP request
- Request replay
- Request export as cURL

## Build
```bash
go mod tidy
wails build
```

## LICENSE
This project is licensed under the MIT License