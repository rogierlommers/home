package homeassistant

import "sync"

type Cache struct {
	id          int
	maxElements int
	mu          sync.RWMutex
	elements    []interface{}
}

func newCache(maxElements int) *Cache {
	return &Cache{
		maxElements: maxElements,
	}
}

func (h *Cache) Add(newElement interface{}) {
	h.mu.Lock()

	if len(h.elements) >= h.maxElements {
		copy(h.elements, h.elements[len(h.elements)-h.maxElements+1:])
		h.elements = h.elements[:h.maxElements-1]
	}
	h.elements = append(h.elements, newElement)

	h.mu.Unlock()
}

func (h *Cache) GetElements() []interface{} {
	return h.elements
}

func (h *Cache) GetElementsReversed() []interface{} {
	var reversed []interface{}

	for i := len(h.elements) - 1; i >= 0; i-- {
		reversed = append(reversed, h.elements[i])
	}

	return reversed
}

func (h *Cache) Count() int {
	return len(h.elements)
}
