package cgoalloc

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

// Allocator is the base interface of cgoalloc- libraries that want to make use of cgoalloc should arrange for their
// methods to accept an Allocator at runtime and use the interface's Malloc/Free to interact with memory.  (cgoalloc.CString,
// cgoalloc.CStringN, and cgoalloc.CBytes provide similar functionality to their C.* equivalents, but accept an Allocator).
//
// Executable packages that want to make use of cgoalloc should initialize one or more implementation of Allocator and use
// them- when finished with an Allocator it's useful to call Destroy.  This will ensure that any C memory pages will be
// deallocated, and it will also attempt to return an error if any Malloc has not been paired with a Free
type Allocator interface {
	// Malloc is equivalent to C.Malloc
	Malloc(size int) unsafe.Pointer
	// Free is equivalent to C.Free
	Free(pointer unsafe.Pointer)
	//
	Destroy() error
}

type DefaultAllocator struct {}

func (a *DefaultAllocator) Malloc(size int) unsafe.Pointer {
	return C.malloc(C.size_t(size))
}

func (a *DefaultAllocator) Free(pointer unsafe.Pointer) {
	C.free(pointer)
}

func (a *DefaultAllocator) Destroy() error {return nil }

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
