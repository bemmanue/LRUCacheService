package lrucache

import (
	"container/heap"
	"container/list"
	"context"
	"fmt"
	"time"
)

type CacheWithTTL struct {
	Cache
	expQueue expirationQueue
	expCheck time.Duration
}

func NewWithTTL(cap int, expCheck time.Duration) (*CacheWithTTL, context.CancelFunc) {
	cache := &CacheWithTTL{
		Cache: Cache{
			cap:   cap,
			data:  make(map[string]*list.Element, cap),
			queue: list.New(),
		},
		expQueue: newExpirationQueue(),
		expCheck: expCheck,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go cache.StartGC(ctx)

	return cache, cancel
}

func (c *CacheWithTTL) StartGC(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("CANCEL")
			return
		default:
			<-time.After(c.expCheck)

			if c.expQueue.Len() == 0 {
				break
			}

			c.mutex.Lock()

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

			c.mutex.Unlock()
		}
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
	c.expQueue = newExpirationQueue()
}

func (c *CacheWithTTL) Add(key string, value any) {
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

func (c *CacheWithTTL) AddWithTTL(key string, value interface{}, ttl time.Duration) {
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

func (c *CacheWithTTL) Get(key string) (any, bool) {
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

func (c *CacheWithTTL) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.data[key]; ok {
		if elem.Value.(*Element).expQueueIndex != -1 {
			heap.Remove(&c.expQueue, elem.Value.(*Element).expQueueIndex)
		}
		c.queue.Remove(elem)
		delete(c.data, key)
	}
}
