# Sorted Map Package

The `sortedMap` package provides a custom data structure `SortedMap` designed to handle JSON objects where key order is significant.

## Problem Statement
Standard Go `map` types are unordered. However, in HTTP fingerprinting and analysis, the order of headers (e.g., in a Client Hello or HTTP request) allows for the identification of specific client implementations (browsers, bots, libraries).

## Implementation
* **Structure**: Maintains a `map[string]interface{}` for O(1) lookups and a slice `[]string` (`Order`) to track insertion sequence.
* **JSON Marshalling**: Implements `MarshalJSON` and `UnmarshalJSON` interfaces. It serializes the data into a custom format containing both the entries and the order array, ensuring that the order allows reconstruction of the original sequence on the frontend or during replay.
