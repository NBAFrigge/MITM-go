# Session Package

Implements in-memory storage for captured HTTP and WebSocket sessions and search logic.

## Public Methods

### `InMemoryStore`

- **`NewInMemoryStore(maxSize int) *InMemoryStore`**
  Creates a store with a maximum size (circular buffer). When the limit is reached, the oldest sessions are removed.

- **`Store(session *sessiondata.Session) error`**
  Saves or updates a session. Manages concurrency via Mutex. If a new session is added that exceeds `maxSize`, it removes the oldest one. Asynchronously notifies all subscribers.

- **`Get(id string) (*sessiondata.Session, error)`**
  Retrieves a single session by ID.

- **`GetAll() []*sessiondata.Session`**
  Returns a copy of all sessions currently in memory, preserving insertion order.

- **`Clear()`**
  Completely empties the store and resets indices.

- **`Subscribe(callback func())`**
  Allows registering a callback function that will be invoked whenever a session is added or updated (used to update the UI via Wails).

- **`Search(opt SearchOptions) ([]*sessiondata.Session, error)`**
  Performs a linear search on saved sessions.
  - `SearchOptions`: Allows filtering by URL, Header Key/Value, Cookie Key/Value, and Body content.
  - Supports both partial matches (case-insensitive substring) and RegEx (if the string is enclosed in `/`).
