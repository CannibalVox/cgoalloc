package cgoalloc

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFixedBlock_TenAllocs(t *testing.T) {
	testAlloc := CreateTestAllocator(t, &DefaultAllocator{})
	alloc, err := CreateFixedBlockAllocator(testAlloc, 160, 8, 8)
	require.NoError(t, err)

	a1 := alloc.Malloc(8)
	a2 := alloc.Malloc(8)
	a3 := alloc.Malloc(8)
	a4 := alloc.Malloc(8)
	a5 := alloc.Malloc(8)
	a6 := alloc.Malloc(8)
	a7 := alloc.Malloc(8)
	a8 := alloc.Malloc(8)
	a9 := alloc.Malloc(8)
	a10 := alloc.Malloc(8)

	alloc.Free(a4)
	alloc.Free(a6)
	alloc.Free(a3)
	alloc.Free(a9)
	alloc.Free(a10)
	alloc.Free(a2)
	alloc.Free(a7)
	alloc.Free(a1)
	alloc.Free(a5)
	alloc.Free(a8)

	allocs, frees := testAlloc.Record()
	require.Len(t, allocs, 1)
	require.Len(t, frees, 0)
	require.Equal(t, 168, allocs[0])

	alloc.Destroy()
	allocs, frees = testAlloc.Record()
	require.Len(t, allocs, 1)
	require.Len(t, frees, 1)
	require.Equal(t, 168, frees[0])
}

func TestFixedBlock_FourPagesUpTwoDown(t *testing.T) {
	testAlloc := CreateTestAllocator(t, &DefaultAllocator{})
	alloc, err := CreateFixedBlockAllocator(testAlloc, 16, 8, 8)
	require.NoError(t, err)
	defer alloc.Destroy()

	a1 := alloc.Malloc(2)
	a2 := alloc.Malloc(2)
	a3 := alloc.Malloc(2)
	a4 := alloc.Malloc(2)
	a5 := alloc.Malloc(2)
	a6 := alloc.Malloc(2)
	a7 := alloc.Malloc(2)

	alloc.Free(a2)
	alloc.Free(a1)
	alloc.Free(a4)
	alloc.Free(a5)
	alloc.Free(a3)

	allocs, frees := testAlloc.Record()
	require.Len(t, allocs, 4)
	require.Len(t, frees, 1)
	require.Equal(t, 24, allocs[0])
	require.Equal(t, 24, allocs[1])
	require.Equal(t, 24, allocs[2])
	require.Equal(t, 24, allocs[3])
	require.Equal(t, 24, frees[0])

	alloc.Free(a6)
	alloc.Free(a7)
}
