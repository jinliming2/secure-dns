package selector

import (
	"sync/atomic"

	"github.com/jinliming2/secure-dns/client/resolver"
)

type clockData struct {
	health int32
}

// Clock return items one by one
type Clock struct {
	clients []*Item

	length int32
	index  int32
}

// Name of selector
func (clock *Clock) Name() string {
	return "Clock"
}

// Add item to list
func (clock *Clock) Add(weight int32, client resolver.DNSClient) {
	clock.clients = append(clock.clients, &Item{
		Client: &client,
		weight: weight,
		data: &clockData{
			health: 10,
		},
	})
	clock.length++
}

// Empty Selector?
func (clock *Clock) Empty() bool {
	return len(clock.clients) == 0
}

// Start set index
func (clock *Clock) Start() {
	clock.index = randomSource.Int31n(clock.length)
}

// Get an item
func (clock *Clock) Get() *Item {
	if clock.length == 0 {
		return nil
	}

	index := clock.index
	i := index + 1
	if i >= clock.length {
		i = 0
	}

	for i != index {
		if atomic.LoadInt32(&clock.clients[i].data.(*clockData).health) > 0 {
			atomic.StoreInt32(&clock.index, i)
			return clock.clients[i]
		}
		i++
		if i >= clock.length {
			i = 0
		}
	}

	i++
	if i >= clock.length {
		i = 0
	}

	atomic.StoreInt32(&clock.index, i)
	return clock.clients[i]
}

// SetHealth of an item
func (clock *Clock) SetHealth(item *Item, score int32) {
	atomic.AddInt32(&item.data.(*clockData).health, score)
}
