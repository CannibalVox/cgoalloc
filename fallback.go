package cgoalloc

import (
	"unsafe"
)

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
	if size > a.fixedBlock.BlockSize() {
		return a.fallback.Malloc(size)
	}

	return a.fixedBlock.Malloc(size)
}

func (a *FallbackAllocator) Free(ptr unsafe.Pointer) {
	if !a.fixedBlock.TryFree(ptr) {
		a.fallback.Free(ptr)
	}
}

func (a *FallbackAllocator) Destroy() {
	a.fixedBlock.Destroy()
	a.fallback.Destroy()
}
