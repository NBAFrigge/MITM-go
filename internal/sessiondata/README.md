# Session Data Package

This package defines the data models and business logic for captured traffic.

## Data Models

* **Session**: The root aggregate root representing a single transaction. Contains `RequestData`, `ResponseData`, `WebSocketData`, TLS info, and timings.
* **RequestData / ResponseData**: normalized structures containing Method, URL, Headers (`SortedMap`), Cookies, Body, and Content-Type.
* **WebSocketData**: Contains the upgrade handshake details, connection state, and a slice of `WebSocketMessage`s.
* **WebSocketMessage**: Represents a single WS frame (Inbound/Outbound, Opcode, Payload, Timestamp).

## Features

* **Diff Engine**: `RequestDifferences` compares two `Session` objects and calculates field-level differences (Added/Removed/Modified headers, body changes, etc.).
* **cURL Export**: `ToCurl` generates a valid cURL command string to replicate the captured request, preserving header order.
* **Replay**: Logic to re-execute a captured request using the proxy's current configuration.
