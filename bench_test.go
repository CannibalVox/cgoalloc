package cgoalloc

import (
	"github.com/stretchr/testify/require"
	"testing"
	"unsafe"
)

func BenchmarkDefaultTemporaryData(b *testing.B) {
	alloc := &DefaultAllocator{}
	defer require.NoError(b, alloc.Destroy())

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
	defer require.NoError(b, alloc.Destroy())

	for i := 0; i < b.N; i++ {
		a := alloc.Malloc(64)
		alloc.Free(a)
	}
}

func BenchmarkArenaTemporaryData(b *testing.B) {
	alloc, err := CreateFixedBlockAllocator(&DefaultAllocator{}, 4096, 64, 8)
	if err != nil {
		b.FailNow()
	}
	defer require.NoError(b, alloc.Destroy())

	for i := 0; i < b.N/2; i++ {
		arena := CreateArenaAllocator(alloc)
		_ = arena.Malloc(64)
		_ = arena.Malloc(64)
		arena.FreeAll()
	}
}

func BenchmarkDefaultGrowShrink(b *testing.B) {
	alloc := &DefaultAllocator{}
	defer require.NoError(b, alloc.Destroy())

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
	defer require.NoError(b, alloc.Destroy())

	ptrs := make([]unsafe.Pointer, b.N, b.N)

	for i := 0; i < b.N; i++ {
		ptrs[i] = alloc.Malloc(8)
	}

	for i := 0; i < b.N; i++ {
		alloc.Free(ptrs[i])
	}
}

func BenchmarkMultilayerTemporaryData(b *testing.B) {
	defAlloc := &DefaultAllocator{}
	lowerLevel, err := CreateFixedBlockAllocator(defAlloc, 256, 8, 8)
	if err != nil {
		b.Fail()
	}
	higherLevel, err := CreateFixedBlockAllocator(defAlloc, 2048, 256, 8)
	if err != nil {
		b.Fail()
	}
	alloc := CreateFallbackAllocator(higherLevel, defAlloc)
	alloc = CreateFallbackAllocator(lowerLevel, alloc)
	defer require.NoError(b, alloc.Destroy())

	for i := 0; i < b.N; i++ {
		size := 4
		if i % 100 == 0 {
			size = 512
		} else if i % 10 == 0 {
			size = 128
		}
		a := alloc.Malloc(size)
		alloc.Free(a)
	}
}

func BenchmarkMultilayerGrowShrink(b *testing.B) {
	defAlloc := &DefaultAllocator{}
	lowerLevel, err := CreateFixedBlockAllocator(defAlloc, 1024*1024, 8, 8)
	if err != nil {
		b.Fail()
	}
	higherLevel, err := CreateFixedBlockAllocator(defAlloc, 32*1024*1024, 256, 8)
	if err != nil {
		b.Fail()
	}
	alloc := CreateFallbackAllocator(higherLevel, defAlloc)
	alloc = CreateFallbackAllocator(lowerLevel, alloc)
	defer require.NoError(b, alloc.Destroy())

	ptrs := make([]unsafe.Pointer, b.N, b.N)
	for i := 0; i < b.N; i++ {
		size := 4
		if i % 1000 == 0 {
			size = 512
		} else if i % 100 == 0 {
			size = 128
		}
		ptrs[i] = alloc.Malloc(size)
	}

	for i := 0; i < b.N; i++ {
		alloc.Free(ptrs[i])
	}
}
