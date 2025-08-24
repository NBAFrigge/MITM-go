package clientHello

import (
	"crypto/tls"
	"sync"
)

type ClientHelloCache struct {
	Cache map[string]*tls.Config
	mu    sync.RWMutex
}

func NewClientHelloCache() *ClientHelloCache {
	return &ClientHelloCache{
		Cache: make(map[string]*tls.Config),
		mu:    sync.RWMutex{},
	}
}

func (c *ClientHelloCache) Get(key []byte) (*tls.Config, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	k := GenerateClientHelloHash(key)

	val, exists := c.Cache[k]
	return val, exists
}

func (c *ClientHelloCache) Set(key []byte, config *tls.Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := GenerateClientHelloHash(key)

	c.Cache[k] = config
}
