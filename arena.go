package cgoalloc

import (
	"errors"
	"unsafe"
)

type ArenaAllocator struct {
	inner Allocator

	allocations []unsafe.Pointer
}

func CreateArenaAllocator(inner Allocator) *ArenaAllocator {
	return &ArenaAllocator{
		inner: inner,
		allocations: make([]unsafe.Pointer, 0, 1),
	}
}

func (a *ArenaAllocator) Malloc(size int) unsafe.Pointer {
	alloc := a.inner.Malloc(size)
	a.allocations = append(a.allocations, alloc)
	return alloc
}

func (a *ArenaAllocator) Free(ptr unsafe.Pointer) {
	allocIndex := -1
	for i := 0; i < len(a.allocations); i++ {
		if a.allocations[i] == ptr {
			allocIndex = i
			break
		}
	}

	if allocIndex < 0 {
		panic("arenaallocator: attempted to free a pointer which had not been allocated with this allocator")
	}

	newEnd := len(a.allocations)-1
	a.allocations[allocIndex] = a.allocations[newEnd]
	a.allocations = a.allocations[:newEnd]

	a.inner.Free(ptr)
}

func (a *ArenaAllocator) FreeAll() {
	for i := 0; i < len(a.allocations); i++ {
		a.inner.Free(a.allocations[i])
	}
	a.allocations = nil
}

func (a *ArenaAllocator) Destroy() error {
	if len(a.allocations) > 0 {
		return errors.New("arenaallocator: attempted to Destroy but not all allocations have been freed")
	}
	return a.inner.Destroy()
}
