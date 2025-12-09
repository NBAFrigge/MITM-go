# Internal Packages

This directory contains the core logic of the application. The architecture is modular to separate responsibilities between proxy management, data parsing, encryption, and storage.

## Package Structure

- **bodyParser**: Handles the decoding and formatting of HTTP payloads (Body).
- **certs**: Manages the internal Public Key Infrastructure (PKI), creating a Root CA and generating "leaf" certificates for each intercepted host.
- **clientHello**: Analyzes raw `ClientHello` TLS packets to extract client cryptographic capabilities (Fingerprinting).
- **headerParser**: Low-level parser to read HTTP headers directly from TCP connection bytes, preserving their order.
- **proxy**: The main engine. Contains network listeners, handlers for HTTP/1.1, HTTP/2 (via hijacking), and WebSockets.
- **session**: In-memory (thread-safe) manager for captured sessions, including search functionality and frontend notifications.
- **sessiondata**: Data structure definitions (Models) for Requests, Responses, and WebSocket Messages, with diffing and export logic.
- **sortedMap**: Custom data structure for handling JSON-serializable maps that maintain key insertion order.
