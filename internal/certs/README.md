# Certs Package

The `certs` package manages the Public Key Infrastructure (PKI) required for the MITM proxy operations. It handles the lifecycle of the Root CA and the on-the-fly generation of leaf certificates for intercepted domains.

## Core Components

### CertCache
The `CertCache` struct is the central manager for certificate operations. It utilizes a thread-safe map (`sync.RWMutex`) to cache generated leaf certificates, minimizing the overhead of RSA key generation for repeated connections to the same host.

### Key Features

1.  **CA Management**:
    * Loads an existing Root CA and Private Key from disk.
    * Generates a new Root CA (RSA 2048-bit) if none exists.
    * Persists the CA certificate and key to the filesystem (`certs/` directory).

2.  **Leaf Certificate Generation**:
    * Generates X.509 certificates dynamically for intercepted hosts.
    * Signs certificates using the loaded Root CA.
    * Populates `Subject Alternative Names` (SANs) with both the DNS name and `www.` subdomain (or IP address if applicable).
    * Sets appropriate Key Usages (`DigitalSignature`, `KeyEncipherment`, `ServerAuth`).

## Constants
* **Bit Size**: 2048 (RSA).
* **Validity**: Generated certificates are valid from 10 minutes in the past to 90 days in the future.
