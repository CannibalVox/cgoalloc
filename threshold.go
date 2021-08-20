package cgoalloc

import (
	"unsafe"
)

type ThresholdAllocator struct {
	sizeThreshold  int
	aboveThreshold Allocator
	belowThreshold Allocator

	allocatedAboveThreshold map[unsafe.Pointer]bool
}

func CreateThresholdAllocator(sizeThreshold int, above Allocator, below Allocator) *ThresholdAllocator {
	return &ThresholdAllocator{
		sizeThreshold:  sizeThreshold,
		aboveThreshold: above,
		belowThreshold: below,

		allocatedAboveThreshold: make(map[unsafe.Pointer]bool),
	}
}

func (a *ThresholdAllocator) Malloc(size int) unsafe.Pointer {
	if size > a.sizeThreshold {
		ptr := a.aboveThreshold.Malloc(size)
		a.allocatedAboveThreshold[ptr] = true
		return ptr
	}

	return a.belowThreshold.Malloc(size)
}

func (a *ThresholdAllocator) Free(ptr unsafe.Pointer) {
	_, ok := a.allocatedAboveThreshold[ptr]
	if ok {
		delete(a.allocatedAboveThreshold, ptr)
		a.aboveThreshold.Free(ptr)
		return
	}

	a.belowThreshold.Free(ptr)
}

func (a *ThresholdAllocator) Destroy() {
	if len(a.allocatedAboveThreshold) > 0 {
		panic("thresholdallocator: attempted to Destroy, but not all allocations had been freed")
	}

	a.aboveThreshold.Destroy()
	a.belowThreshold.Destroy()
}
