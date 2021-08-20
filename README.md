# CGOAlloc

Reducing Malloc/Free traffic to cgo

### Why?

Cgo overhead is a little higher than many are comfortable with (at the time of this writing, a simple call tends to run between 4-6x an equivalent JNI call). Where they really get you, though, is the data marshalling. Each individual call to malloc or free is another cgo call with a 30-50ns overhead.

This library provides an Allocator interface which can be used to provide alternative allocators to C.malloc and C.free.  It also provides a Destroy method, which will clean up any overhead allocated via cgo, as well as make a best-effort to panic if any memory has been allocated and not freed via the destroyed Allocator.  This functionality uses whatever information the allocator in question happens to have available, so it should not be considered definitive.  

More importantly, it provides an allocator `FixedBlockAllocator` which sits on top of another Allocator and allows you to malloc large buffers that are doled out in blocks, amortizing the malloc and free calls across the life of a program.

Also available:

* `DefaultAllocator` - calls cgo for Malloc and Free
* `ThresholdAllocator` - If the malloc size is <= a provided value, use one allocator.  Otherwise, use the other.  Allocations made above the threshold size are stored in a map to enable `Free`. You can use this with a `FixedBlockAllocator` to use the default allocator for large requests.  You could also use several to set up a multi-tiered FBA, I suppose. 
* `ArenaAllocator` - sits on top of another allocator.  Exposes a FreeAll method which will free all memory allocated through the ArenaAllocator.

### Are these thread-safe?

The DefaultAllocator is! And as slow as cgo is, it's still far faster than any locking mechanism in existence, so if you need thread safety, that's what you should use.

### What's the performance like?

In terms of memory overhead, it's kind of bad! I use a lot of maps and slices to track allocated-but-not-freed data.  In terms of speed:

```
BenchmarkDefaultTemporaryData
BenchmarkDefaultTemporaryData-16    	12688918	        96.51 ns/op
BenchmarkFBATemporaryData
BenchmarkFBATemporaryData-16        	100000000	        11.07 ns/op
BenchmarkDefaultGrowShrink
BenchmarkDefaultGrowShrink-16       	11640133	       105.1 ns/op
BenchmarkFBAGrowShrink
BenchmarkFBAGrowShrink-16           	61445820	        34.88 ns/op
```

"It's fine!"
