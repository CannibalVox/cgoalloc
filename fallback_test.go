package cgoalloc

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestThresholdAlloc(t *testing.T) {
	test1 := CreateTestAllocator(t, &DefaultAllocator{})

	test2FBA, err := CreateFixedBlockAllocator(&DefaultAllocator{}, 64, 64, 8)
	if err != nil {
		t.FailNow()
	}
	test2 := CreateTestAllocator(t, test2FBA)

	thresholdAllocator := CreateFallbackAllocator(test2, test1)
	defer thresholdAllocator.Destroy()

	a1 := thresholdAllocator.Malloc(8)
	a2 := thresholdAllocator.Malloc(20)
	b1 := thresholdAllocator.Malloc(68)
	a3 := thresholdAllocator.Malloc(64)
	b2 := thresholdAllocator.Malloc(80)
	b3 := thresholdAllocator.Malloc(100)

	thresholdAllocator.Free(a1)
	thresholdAllocator.Free(b2)
	thresholdAllocator.Free(a2)
	thresholdAllocator.Free(b1)
	thresholdAllocator.Free(a3)
	thresholdAllocator.Free(b3)

	allocs, frees := test1.Record()
	require.Len(t, allocs, 3)
	require.Len(t, frees, 3)
	require.ElementsMatch(t, allocs, []int{68, 80, 100})
	require.ElementsMatch(t, frees, []int{68, 80, 100})

	allocs, frees = test2.Record()
	require.Len(t, allocs, 3)
	require.Len(t, frees, 3)
	require.ElementsMatch(t, allocs, []int{8, 20, 64})
	require.ElementsMatch(t, frees, []int{8, 20, 64})
}
