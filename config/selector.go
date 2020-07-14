package config

// Selectors type
type Selectors string

const (
	// SelectorClock use clock selector
	SelectorClock = Selectors("clock")
	// SelectorRandom use random selector
	SelectorRandom = Selectors("random")
	// SelectorSWRR use Smooth-weighted-round-robin selector
	SelectorSWRR = Selectors("swrr")
	// SelectorWRandom use Weighted-random selector
	SelectorWRandom = Selectors("wrandom")
)