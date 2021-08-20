package cgoalloc

/*
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"sort"
	"unsafe"
)

type Page struct {
	pageStart uintptr
	freeBlocks []unsafe.Pointer
}

type FixedBlockAllocator struct {
	inner Allocator

	allFreeBlocks int
	nextOpenPage int

	pageSize uintptr
	blockSize uintptr
	alignment uintptr
	blocksPerPage int

	pageStarts []uintptr
	pages map[uintptr]*Page
}

func CreateFixedBlockAllocator(inner Allocator, pageSize , blockSize, alignment uintptr) (*FixedBlockAllocator, error) {
	if blockSize % alignment != 0 {
		return nil, errors.New("fixed block allocator: blocksize must be a multiple of alignment")
	}
	if pageSize % blockSize != 0 {
		return nil, errors.New("fixed block allocator: pagesize must be a multiple of blocksize")
	}

	return &FixedBlockAllocator{
		inner: inner,

		allFreeBlocks: 0,
		nextOpenPage: 0,

		pageSize: pageSize,
		blockSize: blockSize,
		alignment: alignment,
		blocksPerPage: int(pageSize/blockSize),

		pages: make(map[uintptr]*Page),
	}, nil
}

func (a *FixedBlockAllocator) Destroy() {
	blocks := a.blocksPerPage * len(a.pages)
	if blocks > a.allFreeBlocks {
		panic("fixedblockallocator: attempted to Destroy, but not all allocations had been freed")
	}

	for _, page := range a.pages {
		a.inner.Free(unsafe.Pointer(page.pageStart))
	}
}

func (a *FixedBlockAllocator) allocatePage() {
	// Allocate page memory
	size := int(a.pageSize+a.alignment)
	pagePtr := a.inner.Malloc(size)

	// Get page bounds & create page
	pageStart := uintptr(pagePtr)
	page := &Page{pageStart: pageStart, freeBlocks: make([]unsafe.Pointer, a.blocksPerPage)}

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

	if a.nextOpenPage > insertIdx {
		a.nextOpenPage = insertIdx
	}
	a.pages[pageStart] = page
}

func (a *FixedBlockAllocator) deallocatePage(page *Page) {
	for i := 0; i < len(a.pageStarts); i++ {
		if a.pageStarts[i] == page.pageStart {
			if a.nextOpenPage > i {
				a.nextOpenPage--
			}
			a.pageStarts = append(a.pageStarts[:i], a.pageStarts[i+1:]...)
			break
		}
	}

	delete(a.pages, page.pageStart)
	a.allFreeBlocks -= a.blocksPerPage
	a.inner.Free(unsafe.Pointer(page.pageStart))
}

func (a *FixedBlockAllocator) Malloc(size uint) unsafe.Pointer {
	if size > uint(a.blockSize) {
		panic("fixed block allocator: requested allocation larger than block size")
	}

	if a.allFreeBlocks == 0 {
		a.allocatePage()
	}

	var block unsafe.Pointer
	for i := a.nextOpenPage; i < len(a.pageStarts); i++ {
		page := a.pages[a.pageStarts[i]]
		if len(page.freeBlocks) > 0 {
			block = page.freeBlocks[0]
			page.freeBlocks = page.freeBlocks[1:]

			break
		}
		if i == a.nextOpenPage {
			a.nextOpenPage++
		}
	}

	if block == nil {
		panic("fixed block allocator: a free block was reported but couldn't be found")
	}

	a.allFreeBlocks--
	return block
}

func (a *FixedBlockAllocator) Free(block unsafe.Pointer) {
	//Find the page this block belongs to
	blockPtr := uintptr(block)
	pageStartIdx := sort.Search(len(a.pageStarts), func(i int) bool {
		start := a.pageStarts[i]
		size := a.pageSize+a.alignment
		end := start+size
		return end > blockPtr
	})

	pageStart := a.pageStarts[pageStartIdx]
	if pageStart > blockPtr {
		panic("fixed block allocator: attempted to free a block not located in an allocated page")
	}
	page := a.pages[pageStart]

	// Return the block
	page.freeBlocks = append(page.freeBlocks, block)

	a.allFreeBlocks++

	totalBlocks := len(a.pages) * a.blocksPerPage
	if len(a.pages) > 1 && a.allFreeBlocks >= (3*totalBlocks/4) && len(page.freeBlocks) >= a.blocksPerPage {
		// We have twice as many blocks as we need & this page is unallocated
		a.deallocatePage(page)
	} else if a.nextOpenPage > pageStartIdx {
		// Malloc our next block from this page
		a.nextOpenPage = pageStartIdx
	}
}
