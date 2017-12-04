package block

import (
	"fmt"
)

const (
	// Size is the fixed size of each block in a buffer
	Size  = 512
	delta = 2
)

type Block interface {
	Address(q int64) (block, offset int)
	ReadAddress(p []byte, block, offset int) int
	WriteAddress(p []byte, block, offset int) int
	Ensure(n int64) error
	Len() (nbytes int64)
	Close() error

	SetLen(nbytes int64)
	Dump() string
}

type Fixed struct {
	Buf    [][Size]byte
	alloc  int64
	nblock int
	len    int64
}

func (b *Fixed) Address(q int64) (block, offset int) {
	return int(q / Size), int(q % Size)
}
func (b *Fixed) SetLen(nbytes int64) {
	b.len = nbytes
}
func (b *Fixed) Len() (nbytes int64) {
	return int64(b.len)
}
func (b *Fixed) Close() error {
	b.Buf = nil
	return nil
}

func (b *Fixed) ReadAddress(p []byte, block, offset int) int {
	return copy(p, b.Buf[block][offset:])
}
func (b *Fixed) WriteAddress(p []byte, block, offset int) int {
	return copy(b.Buf[block][offset:], p)
}

func (b *Fixed) Ensure(n int64) error {
	if n > b.alloc {
		b.Buf = append(b.Buf, make([][Size]byte, n/Size+delta)...)
		b.alloc = int64(len(b.Buf)) * Size
	}
	return nil
}

// Dump is a debugging function that prints blocks in diagnostic form
func (b *Fixed) Dump() (s string) {
	s += ""
	for i := 0; i < len(b.Buf); i++ {
		s += fmt.Sprintf("b[%02d]: %q\n", i, b.Buf[i])
	}
	s += fmt.Sprintf("len: %d\n", b.len)
	s += fmt.Sprintf("alloc: %d\n", b.alloc)
	s += fmt.Sprintf("nblocks: %d\n", b.nblock)
	return s
}
