package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	notBeforeOffset      = -10 * time.Minute
	notAfterOffset       = 3 * 24 * 30 * time.Hour
	maxSerialNumberLimit = uint(128)
	bitSize              = 2048
)

type CertCache struct {
	CACert tls.Certificate
	Cache  map[string]tls.Certificate
	mutex  sync.RWMutex
}

func NewCertCache() *CertCache {
	return &CertCache{
		Cache: make(map[string]tls.Certificate),
	}
}

func (c *CertCache) GenerateCA() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), maxSerialNumberLimit)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"HTTP Debugger CA"},
			CommonName:   "HTTP Debugger Root CA",
		},
		NotBefore:             now.Add(notBeforeOffset),
		NotAfter:              now.Add(notAfterOffset),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}

	c.CACert = tlsCert

	return nil
}

func (c *CertCache) GetHostCert(hostName string, caCert tls.Certificate) (tls.Certificate, error) {
	c.mutex.RLock()
	if cert, exists := c.Cache[hostName]; exists {
		c.mutex.RUnlock()
		return cert, nil
	}
	c.mutex.RUnlock()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if cert, exists := c.Cache[hostName]; exists {
		return cert, nil
	}

	err := c.generateHostCert(hostName, caCert)
	if err != nil {
		return tls.Certificate{}, err
	}

	return c.Cache[hostName], nil
}

func (c *CertCache) generateHostCert(hostName string, caCert tls.Certificate) error {
	ca, err := x509.ParseCertificate(caCert.Certificate[0])
	if err != nil {
		return err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), maxSerialNumberLimit)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"HTTP Debugger"},
			CommonName:   hostName,
		},
		NotBefore:   now.Add(notBeforeOffset),
		NotAfter:    now.Add(notAfterOffset),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	if ip := net.ParseIP(hostName); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, hostName)
		if !strings.HasPrefix(hostName, "www.") {
			template.DNSNames = append(template.DNSNames, "www."+hostName)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, ca, &privateKey.PublicKey, caCert.PrivateKey)
	if err != nil {
		return err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}

	c.Cache[hostName] = tlsCert

	return nil
}

func (c *CertCache) LoadOrGenerateCA(folder, certFile, keyFile string) error {
	err := c.createFolderIfNotExists(folder)
	if err != nil {
		return fmt.Errorf("failed to create folder %s: %w", folder, err)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil {
		c.CACert = cert
		return nil
	}

	err = c.GenerateCA()
	if err != nil {
		return err
	}

	if err := c.saveCertAndKey(cert, certFile, keyFile); err != nil {
		return err
	}

	c.CACert = cert

	return nil
}

func (c *CertCache) saveCertAndKey(cert tls.Certificate, certFile, keyFile string) error {
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}

	privateKey, ok := cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("private key is not RSA")
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: x509Cert.Raw}); err != nil {
		return err
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	return nil
}

func (c *CertCache) checkFolderExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	return nil
}

func (c *CertCache) createFolderIfNotExists(path string) error {
	if err := c.checkFolderExists(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Creating folder %s\n", path)
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create folder %s: %w", path, err)
			}
			return nil
		}
		return err
	}
	return nil
}
