package selector

import (
	"time"

	"github.com/jinliming2/secure-dns/client/resolver"
)

type wRandomInfo struct {
	min    int32
	max    int32
	client *Item
}

// WRandom return items randomally with weight
type WRandom struct {
	clients []wRandomInfo

	length int32
}

// Name of selector
func (wRandom *WRandom) Name() string {
	return "Weighted random"
}

// Add item to list
func (wRandom *WRandom) Add(weight int32, client resolver.DNSClient) {
	wRandom.clients = append(wRandom.clients, wRandomInfo{
		min: wRandom.length,          // [
		max: wRandom.length + weight, // )
		client: &Item{
			Client: &client,
			weight: weight,
		},
	})
	wRandom.length += weight
}

// Empty Selector?
func (wRandom *WRandom) Empty() bool {
	return len(wRandom.clients) == 0
}

// Start set index
func (wRandom *WRandom) Start() {
	randomSource.Seed(time.Now().UnixNano())
}

// Get an item
func (wRandom *WRandom) Get() *Item {
	index := randomSource.Int31n(wRandom.length)
	for _, info := range wRandom.clients {
		if info.min <= index && index < info.max {
			return info.client
		}
	}
	return nil
}

// SetHealth of an item
func (wRandom *WRandom) SetHealth(item *Item, score int32) {}
