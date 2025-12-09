# Header Parser Package

The `headerParser` package provides facilities for parsing HTTP headers directly from raw byte slices.

## Purpose
In standard Go `net/http` handling, the request body is often consumed or the headers are normalized. When hijacking a TCP connection for MITM, access to the raw header block is necessary to preserve the exact order and casing of headers before constructing the internal session representation.

## Implementation
* **ParseHeadersFromRaw**: Reads from a byte slice using `bufio.Reader`.
* **Parsing Logic**: Iterates line-by-line until the CRLF (`\r\n`) delimiter is found. Splits lines into Key-Value pairs and stores them in a `sortedMap.SortedMap` to ensure the original header order is preserved.
