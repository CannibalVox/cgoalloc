package cgoalloc

/*
#include <stdlib.h>
*/
import "C"
import (
	"container/heap"
	"errors"
	"sort"
	"unsafe"
)

// FixedBlockAllocator is an Allocator implementation which reduces cgo.Malloc and cgo.Free traffic by allocating
// entire pages of data at once, and returning pre-assigned block pointers in response to Malloc calls.  Malloc calls
// requesting data buffers larger than the pre-assigned block size will panic.  The Free method does not call C.free-
// instead, the block pointer is simply returned to the allocator to be reused at a later time.
//
// Because the FixedBlockAllocator deals with equally-sized blocks, there is no risk of memory fragmentation. Whenever
// a Malloc is requested, but there are no free block pointers, a new page will be allocated.  Whenever a Free is requested,
// and post-free the page has no assigned block pointers, and fewer than 1/4 of all block pointers are assigned, the
// page will be freed.  Otherwise, Malloc and Free calls made to this Allocator will simply shuffle around block pointers
// with no cgo interaction at all.
type FixedBlockAllocator interface {
	Allocator
	tryFree(ptr unsafe.Pointer) bool
	assignedBlockSize() int
}

type fixedBlockAllocatorImpl struct {
	inner Allocator

	nextPageTicket uint
	allFreeBlocks int

	pageSize uintptr
	blockSize uintptr
	alignment uintptr
	blocksPerPage int

	pageStarts []uintptr
	pages          map[uintptr]*page
	freeBlockQueue pagePQueue
}

// CreateFixedBlockAllocator creates a new FixedBlockAllocator with the provided properties.
// inner - Pages are created using this Allocator
// pageSize - The size of allocated pages, in bytes.  Must be a multiple of blockSize.
// blockSize - The maximum buffer size of requested allocations.  Must be a multiple of alignment.
// alignment - All block pointers will be along this byte alignment.
func CreateFixedBlockAllocator(inner Allocator, pageSize , blockSize, alignment uintptr) (FixedBlockAllocator, error) {
	if blockSize % alignment != 0 {
		return nil, errors.New("fixed block allocator: blocksize must be a multiple of alignment")
	}
	if pageSize % blockSize != 0 {
		return nil, errors.New("fixed block allocator: pagesize must be a multiple of blocksize")
	}

	return &fixedBlockAllocatorImpl{
		inner: inner,

		allFreeBlocks: 0,
		nextPageTicket: 0,

		pageSize: pageSize,
		blockSize: blockSize,
		alignment: alignment,
		blocksPerPage: int(pageSize/blockSize),

		pages: make(map[uintptr]*page),
	}, nil
}

func (a *fixedBlockAllocatorImpl) assignedBlockSize() int { return int(a.blockSize)}

func (a *fixedBlockAllocatorImpl) Destroy() error {
	blocks := a.blocksPerPage * len(a.pages)
	if blocks > a.allFreeBlocks {
		return errors.New("fixedblockallocator: attempted to Destroy, but not all allocations had been freed")
	}

	for _, page := range a.pages {
		a.inner.Free(unsafe.Pointer(page.pageStart))
	}

	return nil
}

func (a *fixedBlockAllocatorImpl) allocatePage() {
	// Allocate page memory
	size := int(a.pageSize+a.alignment)
	pagePtr := a.inner.Malloc(size)

	// Get page bounds & create page
	pageStart := uintptr(pagePtr)
	page := &page{index: -1, pageTicket: a.nextPageTicket, pageStart: pageStart, freeBlocks: make([]unsafe.Pointer, a.blocksPerPage)}
	a.nextPageTicket++

	// Calculate block pointers
	block := unsafe.Pointer((pageStart+a.alignment) - (pageStart % a.alignment))
	for i := 0; i < a.blocksPerPage; i++ {
		page.freeBlocks[i] = block
		block = unsafe.Add(block, C.int(a.blockSize))
	}

	// Add page to allocator
	a.allFreeBlocks += len(page.freeBlocks)

	// Find the insertion point for the new page
	insertIdx := sort.Search(len(a.pageStarts), func(i int) bool {
		return pageStart < a.pageStarts[i]
	})
	a.pageStarts = append(a.pageStarts, 0)
	copy(a.pageStarts[insertIdx+1:],a.pageStarts[insertIdx:])
	a.pageStarts[insertIdx] = pageStart

	// Add to the PQueue
	heap.Push(&a.freeBlockQueue, page)

	a.pages[pageStart] = page
}

func (a *fixedBlockAllocatorImpl) deallocatePage(page *page) {
	for i := 0; i < len(a.pageStarts); i++ {
		if a.pageStarts[i] == page.pageStart {
			a.pageStarts = append(a.pageStarts[:i], a.pageStarts[i+1:]...)
			break
		}
	}

	delete(a.pages, page.pageStart)
	a.freeBlockQueue.Remove(page)

	a.allFreeBlocks -= a.blocksPerPage
	a.inner.Free(unsafe.Pointer(page.pageStart))
}

func (a *fixedBlockAllocatorImpl) Malloc(size int) unsafe.Pointer {
	if size > int(a.blockSize) {
		panic("fixed block allocator: requested allocation larger than block size")
	}

	if a.allFreeBlocks == 0 {
		a.allocatePage()
	}

	page := a.freeBlockQueue.Peek()
	if page == nil || len(page.freeBlocks) == 0 {
		panic("fixed block allocator: a free block was reported but couldn't be found")
	}

	var block unsafe.Pointer
	freeBlockCount := len(page.freeBlocks)
	block = page.freeBlocks[freeBlockCount-1]
	page.freeBlocks = page.freeBlocks[:freeBlockCount-1]
	if len(page.freeBlocks) == 0 {
		_ = heap.Pop(&a.freeBlockQueue)
	}

	a.allFreeBlocks--
	return block
}

func (a *fixedBlockAllocatorImpl) Free(block unsafe.Pointer) {
	if !a.tryFree(block) {
		panic("fixed block allocator: attempted to free a block not located in an allocated page")
	}
}

func (a *fixedBlockAllocatorImpl) tryFree(block unsafe.Pointer) bool {
	pageStartsLen := len(a.pageStarts)
	if pageStartsLen == 0 {
		return false
	}

	//Find the page this block belongs to
	blockPtr := uintptr(block)
	pageStartIdx := sort.Search(pageStartsLen, func(i int) bool {
		start := a.pageStarts[i]
		size := a.pageSize+a.alignment
		end := start+size
		return end > blockPtr
	})

	if pageStartIdx >= pageStartsLen {
		return false
	}

	pageStart := a.pageStarts[pageStartIdx]
	if pageStart > blockPtr {
		return false
	}
	page := a.pages[pageStart]

	// Return the block
	page.freeBlocks = append(page.freeBlocks, block)
	if len(page.freeBlocks) == 1 {
		heap.Push(&a.freeBlockQueue, page)
	}

	a.allFreeBlocks++

	totalBlocks := len(a.pages) * a.blocksPerPage
	if len(a.pages) > 1 && a.allFreeBlocks >= (3*totalBlocks/4) && len(page.freeBlocks) >= a.blocksPerPage {
		// We have twice as many blocks as we need & this page is unallocated
		a.deallocatePage(page)
	}

	return true
}
