package cgoalloc

import (
	"container/heap"
	"unsafe"
)

type Page struct {
	pageTicket uint
	pageStart uintptr
	freeBlocks []unsafe.Pointer

	index int
}

type PagePQueue []*Page

func (pq PagePQueue) Len() int {return len(pq)}

func (pq PagePQueue) Less(i, j int) bool {
	return pq[i].pageTicket > pq[j].pageTicket
}

func (pq PagePQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PagePQueue) Push(x interface{}) {
	item := x.(*Page)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *PagePQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0:n-1]
	item.index = -1
	return item
}

func (pq PagePQueue) Peek() *Page {
	if len(pq) == 0 { return nil }

	return pq[len(pq)-1]
}

func (pq *PagePQueue) Remove(p *Page) {
	if p.index >= 0 {
		heap.Remove(pq, p.index)
	}
}
