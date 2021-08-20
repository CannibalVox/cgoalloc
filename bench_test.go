package cgoalloc

import (
	"testing"
	"unsafe"
)

func BenchmarkDefaultTemporaryData(b *testing.B) {
	alloc := &DefaultAllocator{}
	defer alloc.Destroy()

	for i := 0; i < b.N; i++ {
		a := alloc.Malloc(64)
		alloc.Free(a)
	}
}

func BenchmarkFBATemporaryData(b *testing.B) {
	alloc, err := CreateFixedBlockAllocator(&DefaultAllocator{}, 4096, 64, 8)
	if err != nil {
		b.FailNow()
	}
	defer alloc.Destroy()

	for i := 0; i < b.N; i++ {
		a := alloc.Malloc(64)
		alloc.Free(a)
	}
}

func BenchmarkDefaultGrowShrink(b *testing.B) {
	alloc := &DefaultAllocator{}
	defer alloc.Destroy()

	ptrs := make([]unsafe.Pointer, b.N, b.N)

	for i := 0; i < b.N; i++ {
		ptrs[i] = alloc.Malloc(8)
	}

	for i := 0; i < b.N; i++ {
		alloc.Free(ptrs[i])
	}
}

func BenchmarkFBAGrowShrink(b *testing.B) {
	alloc, err := CreateFixedBlockAllocator(&DefaultAllocator{}, 1024*1024, 8, 8)
	if err != nil {
		b.FailNow()
	}
	defer alloc.Destroy()

	ptrs := make([]unsafe.Pointer, b.N, b.N)

	for i := 0; i < b.N; i++ {
		ptrs[i] = alloc.Malloc(8)
	}

	for i := 0; i < b.N; i++ {
		alloc.Free(ptrs[i])
	}
}
