package clientHello

import (
	"sync"
)

type ClientHelloCache struct {
	Cache map[string]*TLSFingerprint
	mu    sync.RWMutex
}

func NewClientHelloCache() *ClientHelloCache {
	return &ClientHelloCache{
		Cache: make(map[string]*TLSFingerprint),
		mu:    sync.RWMutex{},
	}
}

func (c *ClientHelloCache) Get(key []byte) (*TLSFingerprint, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	k := GenerateClientHelloHash(key)
	val, exists := c.Cache[k]
	return val, exists
}

func (c *ClientHelloCache) Set(key []byte, fingerprint *TLSFingerprint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	k := GenerateClientHelloHash(key)
	c.Cache[k] = fingerprint
}
