# Session Package

The `session` package implements the storage layer for captured traffic.

## InMemoryStore
A concurrent-safe storage engine (`sync.RWMutex`) that keeps sessions in memory.
* **Circular Buffer**: Implements a maximum size limit. When the limit is reached, the oldest sessions are discarded to prevent memory exhaustion.
* **Subscription Model**: Supports a publish-subscribe pattern (`Subscribe`, `notifySubscribers`) to notify the frontend (via Wails runtime) when new data is available.

## Search
Implements the search logic for filtering sessions based on user criteria.
* **Criteria**: URL, Header Keys/Values, Cookies Keys/Values, Body content.
* **Matching**: Supports standard substring matching and Regex matching (if the pattern is enclosed in `/`).
* **Logic**: Uses reflection and helper functions to inspect deeply nested session structures.
