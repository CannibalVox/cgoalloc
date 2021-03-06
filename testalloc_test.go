package cgoalloc

import (
	"github.com/stretchr/testify/require"
	"testing"
	"unsafe"
)

type TestAlloc struct {
	t *testing.T

	inner Allocator

	allocations []int
	frees []int

	allocSizes map[unsafe.Pointer]int
}

func CreateTestAllocator(t *testing.T, inner Allocator) *TestAlloc {
	return &TestAlloc{
		t: t,
		inner: inner,

		allocSizes: make(map[unsafe.Pointer]int),
	}
}

func (a *TestAlloc) Malloc(size int) unsafe.Pointer {
	a.allocations = append(a.allocations, size)

	alloc := a.inner.Malloc(size)
	a.allocSizes[alloc] = size

	return alloc
}

func (a *TestAlloc) recordFree(ptr unsafe.Pointer) {
	size, ok := a.allocSizes[ptr]
	require.True(a.t, ok)

	delete(a.allocSizes, ptr)
	a.frees = append(a.frees, size)
}

func (a *TestAlloc) assignedBlockSize() int {
	fba, ok := a.inner.(FixedBlockAllocator)
	require.True(a.t, ok, "testalloc: used testalloc as a fixedbufferallocator but it isn't wrapping a fixedbufferallocator")
	return fba.assignedBlockSize()
}

func (a *TestAlloc) tryFree(ptr unsafe.Pointer) bool {
	fba, ok := a.inner.(FixedBlockAllocator)
	require.True(a.t, ok, "testalloc: used testalloc as a fixedbufferallocator but it isn't wrapping a fixedbufferallocator")

	freed := fba.tryFree(ptr)
	if freed {
		a.recordFree(ptr)
	}
	return freed
}

func (a *TestAlloc) Free(ptr unsafe.Pointer) {
	a.recordFree(ptr)
	a.inner.Free(ptr)
}

func (a *TestAlloc) Destroy() error {
	require.Len(a.t, a.allocSizes, 0)
	return a.inner.Destroy()
}

func (a *TestAlloc) Record() (allocs []int, frees []int) {
	return a.allocations, a.frees
}
