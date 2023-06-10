package lrucache

import (
	"container/heap"
	"container/list"
	"time"
)

type CacheWithTTL2 struct {
	Cache
	expQueue expirationQueue
}

func NewWithTTL2(cap int) *CacheWithTTL2 {
	return &CacheWithTTL2{
		Cache: Cache{
			cap:   cap,
			data:  make(map[string]*list.Element, cap),
			queue: list.New(),
		},
		expQueue: newExpirationQueue(),
	}
}

func (c *CacheWithTTL2) UpdateExpirations() {
	if c.expQueue.Len() == 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check and remove expired elements
	for c.expQueue.Len() > 0 {
		last := c.expQueue[0]

		if last.Value.(*Element).expiresAt.Before(time.Now()) {
			delete(c.data, last.Value.(*Element).key)
			c.queue.Remove(last)
			heap.Pop(&c.expQueue)
		} else {
			break
		}
	}
}

func (c *CacheWithTTL2) Cap() int {
	return c.cap
}

func (c *CacheWithTTL2) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.data)
}

func (c *CacheWithTTL2) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*list.Element, c.cap)
	c.queue = list.New()
	c.expQueue = newExpirationQueue()
}

func (c *CacheWithTTL2) Add(key string, value any) {
	c.UpdateExpirations()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// if element already exists just update element position in queue
	if elem, ok := c.data[key]; ok {
		elem.Value = &Element{
			key:   key,
			value: value,
		}
		c.queue.MoveToFront(elem)
		return
	}

	// if cache is full displace the value that was not requested the most
	if c.queue.Len() == c.cap {
		last := c.queue.Back()
		delete(c.data, last.Value.(*Element).key)
		c.queue.Remove(last)
	}

	// add new element
	newElem := c.queue.PushFront(&Element{
		key:           key,
		value:         value,
		expQueueIndex: -1,
	})
	c.data[key] = newElem
}

func (c *CacheWithTTL2) AddWithTTL(key string, value interface{}, ttl time.Duration) {
	c.UpdateExpirations()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// update element if it already exists
	if elem, ok := c.data[key]; ok {
		elem.Value.(*Element).expiresAt = time.Now().Add(ttl)

		if elem.Value.(*Element).expQueueIndex == -1 {
			c.expQueue.Push(elem)
		} else {
			heap.Fix(&c.expQueue, elem.Value.(*Element).expQueueIndex)
		}

		c.queue.MoveToFront(elem)
		return
	}

	// if cache is full displace the value that was not requested the most
	if c.queue.Len() == c.cap {
		last := c.queue.Back()
		delete(c.data, last.Value.(*Element).key)
		c.queue.Remove(last)
	}

	// add new element
	newElem := c.queue.PushFront(&Element{
		key:           key,
		value:         value,
		expiresAt:     time.Now().Add(ttl),
		expQueueIndex: -1,
	})
	c.data[key] = newElem
	c.expQueue.Push(newElem)
}

func (c *CacheWithTTL2) Get(key string) (any, bool) {
	c.UpdateExpirations()

	c.mutex.RLock()
	elem, ok := c.data[key]
	c.mutex.RUnlock()

	// update position in queue if element exists
	if ok {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.queue.MoveToFront(elem)
		return elem.Value.(*Element).value, true
	}

	return nil, false
}

func (c *CacheWithTTL2) Remove(key string) {
	c.UpdateExpirations()

	c.mutex.RLock()
	elem, ok := c.data[key]
	c.mutex.RUnlock()

	if ok {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		if elem.Value.(*Element).expQueueIndex != -1 {
			heap.Remove(&c.expQueue, elem.Value.(*Element).expQueueIndex)
		}
		c.queue.Remove(elem)
		delete(c.data, key)
	}
}
