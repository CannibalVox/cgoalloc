package cgoalloc

import (
	"unsafe"
)

// FallbackAllocator is an Allocator implementation which accepts a FixedBlockAllocator and sends all Malloc calls which
// can fit in the FBA's block size to that FixedBlockAllocator.  All other calls are sent to a fallback allocator.
type FallbackAllocator struct {
	fixedBlock FixedBlockAllocator
	fallback Allocator
}

func CreateFallbackAllocator(fixedBlock FixedBlockAllocator, fallback Allocator) *FallbackAllocator {
	return &FallbackAllocator{
		fixedBlock: fixedBlock,
		fallback: fallback,
	}
}

func (a *FallbackAllocator) Malloc(size int) unsafe.Pointer {
	if size > a.fixedBlock.assignedBlockSize() {
		return a.fallback.Malloc(size)
	}

	return a.fixedBlock.Malloc(size)
}

func (a *FallbackAllocator) Free(ptr unsafe.Pointer) {
	if !a.fixedBlock.tryFree(ptr) {
		a.fallback.Free(ptr)
	}
}

func (a *FallbackAllocator) Destroy() error {
	err := a.fixedBlock.Destroy()
	if err != nil { return err }
	return a.fallback.Destroy()
}
