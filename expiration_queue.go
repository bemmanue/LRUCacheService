package lrucache

import (
	"container/heap"
	"container/list"
)

type expirationQueue []*list.Element

func newExpirationQueue() expirationQueue {
	var q expirationQueue = make([]*list.Element, 0)
	heap.Init(&q)
	return q
}

func (q expirationQueue) Len() int {
	return len(q)
}

func (q expirationQueue) Less(i, j int) bool {
	return q[i].Value.(Element).expiresAt.Before(q[j].Value.(Element).expiresAt)
}

func (q expirationQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].Value.(*Element).expQueueIndex = i
	q[j].Value.(*Element).expQueueIndex = j
}

func (q *expirationQueue) Push(x any) {
	elem := x.(*list.Element)
	elem.Value.(*Element).expQueueIndex = len(*q)
	*q = append(*q, x.(*list.Element))
}

func (q *expirationQueue) Pop() any {
	old := *q
	n := len(old)
	x := old[n-1]
	x.Value.(*Element).expQueueIndex = -1
	*q = old[0 : n-1]
	return x
}
