package cgoalloc

import (
	"errors"
	"unsafe"
)

// ArenaAllocator is an Allocator implementation which accepts an Allocator object and passes Malloc and Free calls to
// the underlying Allocator.  However, it exposes a FreeAll method which will instantly free any Malloc calls which have
// been proxied through the ArenaAllocator.  Because FreeAll requires all Malloc calls to be tracked in the ArenaAllocator,
// and because that tracking is optimized for FreeAll speed, direct calls to Free have O(N) time and are not recommended.
//
// ArenaAllocator is intended to be spun up temporarily for a flurry of malloc activity that then needs to be undone
// at the end.
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

// FreeAll calls Free for every pointer allocated but not freed through this ArenaAllocator
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
