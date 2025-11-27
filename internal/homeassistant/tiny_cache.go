package homeassistant

import "sync"

type Cache struct {
	maxElements int
	mu          sync.RWMutex
	elements    []any
}

func newCache(maxElements int) *Cache {
	return &Cache{
		maxElements: maxElements,
	}
}

func (h *Cache) Add(newElement any) {
	h.mu.Lock()

	if len(h.elements) >= h.maxElements {
		copy(h.elements, h.elements[len(h.elements)-h.maxElements+1:])
		h.elements = h.elements[:h.maxElements-1]
	}
	h.elements = append(h.elements, newElement)

	h.mu.Unlock()
}

func (h *Cache) GetElements() []any {
	return h.elements
}

func (h *Cache) GetElementsReversed() []any {
	var reversed []any

	for i := len(h.elements) - 1; i >= 0; i-- {
		reversed = append(reversed, h.elements[i])
	}

	return reversed
}

func (h *Cache) Count() int {
	return len(h.elements)
}
