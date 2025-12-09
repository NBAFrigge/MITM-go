# Certs Package

Manages the dynamic generation of X.509 certificates required for SSL/TLS interception.

## Public Methods

### `CertCache`

Keeps generated certificates in memory to avoid the computational cost of RSA key generation for every connection.

- **`NewCertCache() *CertCache`**
  Creates a new thread-safe cache (`sync.RWMutex`).

- **`LoadOrGenerateCA(folder, certFile, keyFile string) error`**
  Attempts to load a CA certificate/key pair from disk. If they don't exist, it generates a new Root CA (RSA 2048 bit), saves it to disk, and loads it into memory.

- **`GenerateCA() error`**
  Programmatically generates a self-signed Root CA certificate valid for approximately 3 months. Sets appropriate `KeyUsage` flags for certificate signing (CertSign).

- **`GetHostCert(hostName string, caCert tls.Certificate) (tls.Certificate, error)`**
  Thread-safe method to obtain a leaf certificate for a specific domain.
  1. Checks the cache for reading (`RLock`).
  2. If absent, acquires the write lock (`Lock`).
  3. Generates a new certificate signed by the loaded CA.
  4. The certificate includes SANs (Subject Alternative Names) for the domain and, if applicable, for the `www` subdomain.

## Constants

- Algorithm: RSA 2048 bit.
- Default Validity: From -10 minutes (for clock skew) to +90 days.
