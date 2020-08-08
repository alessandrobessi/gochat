package activeclients

import (
	"sync"

	"github.com/alessandrobessi/gochat/src/pkg/types"
)

// ActiveClients is a struct to keep track of the active clients
type ActiveClients struct {
	Map map[string]types.Client
	mux sync.Mutex // to avoid race conditions
}

// HasKey returns true if the key is in the map; false otherwise
func (c *ActiveClients) HasKey(key string) bool {
	c.mux.Lock()
	if _, ok := c.Map[key]; ok {
		c.mux.Unlock()
		return true
	}
	c.mux.Unlock()
	return false
}

// GetMap returns a pointer to the map
func (c *ActiveClients) GetMap() *map[string]types.Client {
	return &c.Map
}

// Count returns the number of the active clients
func (c *ActiveClients) Count() int {
	numActiveClients := 0
	for _, client := range c.Map {
		if client.IsActive == true {
			numActiveClients++
		}
	}
	return numActiveClients
}

// CleanUp deletes the non-active clients
func (c *ActiveClients) CleanUp() {
	c.mux.Lock()
	for _, client := range c.Map {
		if client.IsActive == false {
			c.DeleteClient(client.Name)
		}
	}
	c.mux.Unlock()
}

// AddClient adds a client to the map
func (c *ActiveClients) AddClient(key string, value types.Client) {
	c.mux.Lock()
	c.Map[key] = value
	c.mux.Unlock()
}

// DeleteClient deletes a client from the map
func (c *ActiveClients) DeleteClient(key string) {
	c.mux.Lock()
	delete(c.Map, key)
	c.mux.Unlock()
}
