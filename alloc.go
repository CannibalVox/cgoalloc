package cgoalloc

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

type Allocator interface {
	Malloc(size int) unsafe.Pointer
	Free(pointer unsafe.Pointer)
	Destroy()
}

type DefaultAllocator struct {}

func (a *DefaultAllocator) Malloc(size int) unsafe.Pointer {
	return C.malloc(C.size_t(size))
}

func (a *DefaultAllocator) Free(pointer unsafe.Pointer) {
	C.free(pointer)
}

func (a *DefaultAllocator) Destroy() {}

func CString(allocator Allocator, str string) *C.char {
	ptr := allocator.Malloc(len(str)+1)
	ptrArray := (*[1<<30]byte)(ptr)
	copy(ptrArray[:], str)
	ptrArray[len(str)] = 0
	return (*C.char)(ptr)
}

func CBytes(allocator Allocator, b []byte) unsafe.Pointer {
	ptr := allocator.Malloc(len(b))
	ptrArray := (*[1<<30]byte)(ptr)
	copy(ptrArray[:], b)
	return ptr
}
