package selector

import (
	"sync/atomic"

	"github.com/jinliming2/secure-dns/client/resolver"
)

type sWrrData struct {
	currentWeight int32
}

// SWrr implemented smooth weighted round robin
type SWrr struct {
	clients []*Item

	length int32
	index  int32
}

// Name of selector
func (sWrr *SWrr) Name() string {
	return "Smooth weighted round robin"
}

// Add item to list
func (sWrr *SWrr) Add(weight int32, client resolver.DNSClient) {
	sWrr.clients = append(sWrr.clients, &Item{
		Client: &client,
		weight: weight,
		data: &sWrrData{
			currentWeight: 0,
		},
	})
	sWrr.length++
}

// Empty Selector?
func (sWrr *SWrr) Empty() bool {
	return sWrr.length == 0
}

// Start set index
func (sWrr *SWrr) Start() {
	if sWrr.Empty() {
		return
	}
	count := randomSource.Int31n(sWrr.length)
	for i := int32(0); i < count; i++ {
		sWrr.Get()
	}
}

// Get an item
func (sWrr *SWrr) Get() *Item {
	if sWrr.Empty() {
		return nil
	}

	var (
		total int32
		best  *Item
	)

	for _, item := range sWrr.clients {
		atomic.AddInt32(&item.data.(*sWrrData).currentWeight, item.weight)
		total += item.weight

		if best == nil || atomic.LoadInt32(&item.data.(*sWrrData).currentWeight) > atomic.LoadInt32(&best.data.(*sWrrData).currentWeight) {
			best = item
		}
	}

	if best != nil {
		atomic.AddInt32(&best.data.(*sWrrData).currentWeight, -total)
	}

	return best
}

// SetHealth of an item
func (sWrr *SWrr) SetHealth(item *Item, score int32) {
	n := atomic.AddInt32(&item.data.(*sWrrData).currentWeight, score)
	if n < 1 {
		atomic.StoreInt32(&item.data.(*sWrrData).currentWeight, 1)
	} else if n > item.weight {
		atomic.StoreInt32(&item.data.(*sWrrData).currentWeight, item.weight)
	}
}
