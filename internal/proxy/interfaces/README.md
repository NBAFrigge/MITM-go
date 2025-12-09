# Interfaces Package

This package defines the core interfaces used for dependency injection within the proxy architecture to prevent circular dependencies and facilitate testing.

## Interfaces

- **SessionStore**: Defines methods for storing, retrieving, and searching traffic sessions (`Store`, `GetAll`, `Get`).
- **Logger**: Defines the contract for logging request/response lifecycles and errors.
