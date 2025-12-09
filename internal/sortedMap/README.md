# SortedMap Package

Implements a hybrid data structure (Map + Slice) to handle JSON objects where key order is significant.

## Public Methods

### `SortedMap` Struct

Maintains two fields: `Entries` (standard map for O(1) access) and `Order` (slice of strings for order).

- **`New() *SortedMap`**: Initializes the map and slice.
- **`Put(key string, value interface{})`**: Inserts or updates a value. If the key is new, it is appended to the `Order` slice.
- **`Get(key string) (interface{}, bool)`**: Retrieves a value from the internal map.
- **`Delete(key string)`**: Removes from the map and linearly searches the `Order` slice to remove the key, maintaining the order of other elements.
- **`Keys() []string`**: Returns keys in exact insertion order.
- **`Equal(other *SortedMap) bool`**: Compares two SortedMaps verifying both value equality and key order equality.
- **`MarshalJSON() ([]byte, error)`** / **`UnmarshalJSON(data []byte) error`**: Custom implementation of `json.Marshaler` and `json.Unmarshaler` interfaces. Serializes the object into a specific format explicitly including the `order` field in the resulting JSON, allowing the frontend to visually reconstruct the correct order.
