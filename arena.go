package cgoalloc

import (
	"unsafe"
)

type ArenaAllocator struct {
	inner Allocator

	allocations map[unsafe.Pointer]bool
}

func CreateArenaAllocator(inner Allocator) *ArenaAllocator {
	return &ArenaAllocator{
		inner: inner,

		allocations: make(map[unsafe.Pointer]bool),
	}
}

func (a *ArenaAllocator) Malloc(size int) unsafe.Pointer {
	alloc := a.inner.Malloc(size)
	a.allocations[alloc] = true
	return alloc
}

func (a *ArenaAllocator) Free(ptr unsafe.Pointer) {
	_, ok := a.allocations[ptr]
	if !ok {
		panic("arenaallocator: attempted to free a pointer which had not been allocated with this allocator")
	}
	delete(a.allocations, ptr)

	a.inner.Free(ptr)
}

func (a *ArenaAllocator) FreeAll() {
	for ptr := range a.allocations {
		a.inner.Free(ptr)
	}
	a.allocations = make(map[unsafe.Pointer]bool)
}

func (a *ArenaAllocator) Destroy() {
	if len(a.allocations) > 0 {
		panic("arenaallocator: attempted to Destroy but not all allocations have been freed")
	}
	a.inner.Destroy()
}
