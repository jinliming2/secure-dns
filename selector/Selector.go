package selector

import (
	"math/rand"
	"time"

	"github.com/jinliming2/secure-dns/client/resolver"
)

var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// Item is an item in Selector
type Item struct {
	Client *resolver.DNSClient
	weight int32

	data interface{}
}

// Selector implemented round robins
type Selector interface {
	Name() string
	Add(weight int32, client resolver.DNSClient)
	Empty() bool
	Start()
	Get() *Item
	SetHealth(item *Item, score int32)
}
