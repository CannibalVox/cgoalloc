package cgoalloc

import (
	"container/heap"
	"unsafe"
)

type page struct {
	pageTicket uint
	pageStart uintptr
	freeBlocks []unsafe.Pointer

	index int
}

type pagePQueue []*page

func (pq pagePQueue) Len() int {return len(pq)}

func (pq pagePQueue) Less(i, j int) bool {
	return pq[i].pageTicket > pq[j].pageTicket
}

func (pq pagePQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *pagePQueue) Push(x interface{}) {
	item := x.(*page)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *pagePQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0:n-1]
	item.index = -1
	return item
}

func (pq pagePQueue) Peek() *page {
	if len(pq) == 0 { return nil }

	return pq[len(pq)-1]
}

func (pq *pagePQueue) Remove(p *page) {
	if p.index >= 0 {
		heap.Remove(pq, p.index)
	}
}
