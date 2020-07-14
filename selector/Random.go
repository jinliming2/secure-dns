package selector

import (
	"time"

	"github.com/jinliming2/secure-dns/client/resolver"
)

// Random return items randomally
type Random struct {
	clients []*Item

	length int
}

// Name of selector
func (random *Random) Name() string {
	return "Random"
}

// Add item to list
func (random *Random) Add(weight int32, client resolver.DNSClient) {
	random.clients = append(random.clients, &Item{
		Client: &client,
		weight: weight,
	})
	random.length++
}

// Empty Selector?
func (random *Random) Empty() bool {
	return len(random.clients) == 0
}

// Start set index
func (random *Random) Start() {
	randomSource.Seed(time.Now().UnixNano())
}

// Get an item
func (random *Random) Get() *Item {
	return random.clients[randomSource.Intn(random.length)]
}

// SetHealth of an item
func (random *Random) SetHealth(item *Item, score int32) {}
