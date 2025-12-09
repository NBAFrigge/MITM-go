# Header Parser Package

Utility package for low-level parsing of HTTP headers.

## Public Methods

- **`ParseHeadersFromRaw(rawRequest []byte) (*sortedMap.SortedMap, error)`**
  Reads a byte slice containing a raw HTTP request.
  - Uses `bufio.Reader` to read line by line.
  - Ignores the initial Request Line (Method/URL).
  - Processes each subsequent line as a `Key: Value` pair.
  - Stops at the empty line (`\r\n`) separating headers and body.
  - **Important**: Saves results into a `sortedMap.SortedMap`. This is crucial because standard Go maps do not guarantee order, while header order is fundamental for fingerprinting analysis and correct request replication.
