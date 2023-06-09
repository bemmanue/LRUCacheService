package lrucache

import (
	"container/heap"
	"container/list"
	"time"
)

type CacheWithTTL struct {
	*Cache
	expQueue expirationQueue
	expCheck time.Duration
}

func NewWithTTL(cap int, expCheck time.Duration) *CacheWithTTL {
	cache := &CacheWithTTL{
		Cache:    New(cap),
		expQueue: newExpirationQueue(),
		expCheck: expCheck,
	}

	go cache.GC()

	return cache
}

func (c *CacheWithTTL) GC() {
	for {
		<-time.After(c.expCheck)

		if c.expQueue.Len() == 0 {
			continue
		}

		c.mutex.Lock()

		// check and remove expired elements
		for c.expQueue.Len() > 0 {
			last := c.expQueue[0]

			if last.Value.(Element).expiresAt.Before(time.Now()) {
				delete(c.data, last.Value.(Element).key)
				c.queue.Remove(last)
				heap.Pop(&c.expQueue)
			} else {
				break
			}
		}

		c.mutex.Unlock()
	}
}

func (c *CacheWithTTL) Cap() int {
	return c.cap
}

func (c *CacheWithTTL) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.data)
}

func (c *CacheWithTTL) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*list.Element, c.cap)
	c.queue = list.New()
}

func (c *CacheWithTTL) Add(key string, value any) {
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

func (c *CacheWithTTL) AddWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// if element already exists do nothing
	if _, ok := c.data[key]; ok {
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
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
	})
	c.data[key] = newElem
	c.expQueue.Push(newElem)
}

func (c *CacheWithTTL) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if elem, ok := c.data[key]; ok {
		c.queue.MoveToFront(elem)
		return elem.Value.(Element).value, true
	}

	return nil, false
}

func (c *CacheWithTTL) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.data[key]; ok {
		c.queue.Remove(elem)
		delete(c.data, key)
	}
}
