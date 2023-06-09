package lrucache

import (
	"container/list"
	"sync"
)

type Element struct {
	key   string
	value any
}

type Cache struct {
	cap   int
	data  map[string]*list.Element
	mutex sync.RWMutex
	queue *list.List
}

func New(cap int) *Cache {
	return &Cache{
		cap:   cap,
		data:  make(map[string]*list.Element, cap),
		queue: list.New(),
	}
}

func (c *Cache) Cap() int {
	return c.cap
}

func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*list.Element, c.cap)
	c.queue = list.New()
}

func (c *Cache) Add(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// if element already exists just update element position in queue
	if elem, ok := c.data[key]; ok {
		elem.Value = Element{key: key, value: value}
		c.queue.MoveToFront(elem)
		return
	}

	// if cache is full displace the value that was not requested the most
	if c.queue.Len() == c.cap {
		last := c.queue.Back()
		delete(c.data, last.Value.(Element).key)
		c.queue.Remove(last)
	}

	// add new element
	newElem := c.queue.PushFront(Element{
		key:   key,
		value: value,
	})
	c.data[key] = newElem
}

func (c *Cache) Get(key string) (any, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// get value and update position in queue
	if elem, ok := c.data[key]; ok {
		c.queue.MoveToFront(elem)
		return elem.Value.(Element).value, true
	}

	return nil, false
}

func (c *Cache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.data[key]; ok {
		c.queue.Remove(elem)
		delete(c.data, key)
	}
}
