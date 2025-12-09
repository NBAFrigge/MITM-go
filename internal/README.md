# Internal Packages

This directory contains the private library code for the HTTP Debugger application. These packages encompass the core logic for the MITM (Man-in-the-Middle) proxy, TLS handling, traffic interception, session management, and data parsing.

## Directory Structure

* **bodyParser**: Utilities for handling HTTP request and response bodies, including decompression and formatting.
* **certs**: Certificate Authority (CA) management and dynamic generation of leaf certificates for intercepted hosts.
* **clientHello**: Parsing and caching of TLS Client Hello messages to support TLS fingerprinting and signature replication.
* **headerParser**: Low-level parsing of HTTP headers from raw byte streams.
* **proxy**: The core proxy server implementation, including connection wrappers, request handlers (HTTP, HTTPS/MITM, WebSocket), and utilities.
* **session**: In-memory storage and management of captured traffic sessions, including search functionality.
* **sessiondata**: Data structures representing captured HTTP and WebSocket sessions, including logic for diffing and cURL export.
* **sortedMap**: A custom map data structure implementation that preserves key insertion order, essential for maintaining HTTP header integrity.
