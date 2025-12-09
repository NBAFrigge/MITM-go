# Types Package

This package holds shared struct definitions used across the proxy subsystems.

## Config

The `Config` struct is the primary configuration object passed to handlers.

* `SessionStore`: Interface to the storage engine.
* `Logger`: Interface to the logging engine.
* `HTTPClient`: The configured upstream client.
* `CACert`: The Root Certificate used for signing.
* `Mutex`: A global lock for thread-safe operations on shared configuration.
