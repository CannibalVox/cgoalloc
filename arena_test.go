package cgoalloc

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestArena_FreeAll(t *testing.T) {
	testAlloc := CreateTestAllocator(t, &DefaultAllocator{})
	alloc := CreateArenaAllocator(testAlloc)
	defer require.NoError(t, alloc.Destroy())

	_ = alloc.Malloc(8)
	_ = alloc.Malloc(12)
	_ = alloc.Malloc(16)

	alloc.FreeAll()

	allocs, frees := testAlloc.Record()
	require.Len(t, allocs, 3)
	require.Len(t, frees, 3)

	require.ElementsMatch(t, allocs, []int{8, 12, 16})
	require.ElementsMatch(t, frees, []int{8, 12, 16})
}

func TestArena_PreFreeOne(t *testing.T) {
	testAlloc := CreateTestAllocator(t, &DefaultAllocator{})
	alloc := CreateArenaAllocator(testAlloc)
	defer require.NoError(t, alloc.Destroy())

	a1 := alloc.Malloc(8)
	_ = alloc.Malloc(12)
	_ = alloc.Malloc(16)

	alloc.Free(a1)
	alloc.FreeAll()

	allocs, frees := testAlloc.Record()
	require.Len(t, allocs, 3)
	require.Len(t, frees, 3)

	require.ElementsMatch(t, allocs, []int{8, 12, 16})
	require.ElementsMatch(t, frees, []int{8, 12, 16})
}
