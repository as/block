package main

import (
	"bytes"
	"fmt"
	"io"
)

func msg(fm string, i ...interface{}) {
	fmt.Printf(fm, i...)
	fmt.Println()
}

const (
	Size  = 10
	Delta = 10
)

type Buffer struct {
	Buf   [][Size]byte
	p     int
	bp    int
	bn    int
	len   int
	alloc int
}

func main() {
	bl := &Buffer{}
	bl.WriteAt([]byte("g"), 0)
	bl.WriteAt([]byte("o"), 1)
	bl.Insert([]byte("@"), 0)
	bl.Delete(0, 1)
	fmt.Println(bl.String())
}

func (b *Buffer) String() string {
	buf := new(bytes.Buffer)
	lim := &io.LimitedReader{b, int64(b.len)}
	at, _ := b.Seek(0, 1)
	b.Seek(0, 0)
	io.Copy(buf, lim)
	b.Seek(at, 0)
	return buf.String()
}
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekCurrent {
		offset += int64(b.p)
	} else if whence == io.SeekEnd {
		offset = int64(b.len) + offset
	}
	if offset < 0 {
		return int64(b.p), io.ErrUnexpectedEOF
	}
	b.p, b.bn, b.bp = b.findRegion(offset)
	return int64(b.p), nil
}
func (b *Buffer) Dump() (s string) {
	s += ""
	for i := 0; i < len(b.Buf); i++ {
		s += fmt.Sprintf("b[%02d]: %q\n", i, b.Buf[i])
	}
	s += fmt.Sprintf("p: %d\n", b.p)
	s += fmt.Sprintf("bp: %d\n", b.bp)
	s += fmt.Sprintf("bn: %d\n", b.bn)
	s += fmt.Sprintf("len: %d\n", b.len)
	s += fmt.Sprintf("alloc: %d\n", b.alloc)
	return s
}
func (b *Buffer) findRegion(at int64) (p, bn, bp int) {
//	defer func() {
//		fmt.Printf("findregion p/bn/bp: %d/%d/%d\n", p, bn, bp)
//	}()
	return int(at), int(at / Size), int(at % Size)
}
func (b *Buffer) Insert(p []byte, at int64) (n int, err error) {
	growth := b.len - int(at) + len(p)
	b.len += growth
	b.ensure(b.len + growth)
	tmp := make([]byte, Size)
	defer func(at int64) {
		n, err = b.WriteAt(p, at)
	}(at)
	at += int64(len(p))
	for int(at)-len(p) < b.len {
		defer func(at int64) {
			n, err = b.ReadAt(tmp, int64(at))
			n, err = b.WriteAt(tmp[:n], at+int64(len(p)))
		}(at - int64(len(p)))
		at += int64(len(tmp))
	}
	return
}

func (b *Buffer) Delete(q0, q1 int64) (n int, err error) {
	tmp := make([]byte, Size)
	dx := q1 - q0
	b.len -= int(dx)
	for at := int64(0); at+q0 != int64(b.len); {
		n, err = b.ReadAt(tmp, q1+at)
		n, err = b.WriteAt(tmp, q0+at)
		at += int64(n)
	}
	return
}

func (b *Buffer) WriteAt(p []byte, at int64) (n int, err error) {
	if at > int64(b.len) {
		return 0, io.ErrUnexpectedEOF
	}

	x, y, z := b.p, b.bn, b.bp

	b.p, b.bn, b.bp = b.findRegion(at)
	n, err = b.Write(p)

	b.p, b.bn, b.bp = x, y, z
	return n, err
}

func (b *Buffer) ReadAt(p []byte, at int64) (n int, err error) {
	if at > int64(b.len) {
		return 0, io.ErrUnexpectedEOF
	}

	x, y, z := b.p, b.bn, b.bp

	b.p, b.bn, b.bp = b.findRegion(at)
	n, err = b.Read(p)

	b.p, b.bn, b.bp = x, y, z
	return n, err
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.ensure(b.p + len(p))
	for n < len(p) {
		if len(b.Buf) <= b.bn {
			panic("box phase error")
		}
		m := copy(b.Buf[b.bn][b.bp:], p[n:])
		n += m
		b.bp += m
		b.p += m
		//p = p[n:]
		if b.bp > Size {
			panic("b.bp > Size")
		}
		if b.bp == Size {
			b.bp = 0
			b.bn++
		}
	}
	if b.p > b.len {
		b.len = b.p
	}
	return
}
func (b *Buffer) Read(p []byte) (n int, err error) {
	for {
		if b.bp == Size {
			b.bp = 0
			b.bn++
		}
		if b.p == b.len {
			return n, io.EOF
		}
		m := copy(p[n:], b.Buf[b.bn][b.bp:])
		if m == 0 {
			return n, nil
		}
		n += m
		b.bp += m
		b.p += m
		//p = p[n:]
	}
}
func (b *Buffer) ensure(n int) {
	if n > b.alloc {
		b.Buf = append(b.Buf, make([][Size]byte, n/Size+Delta)...)
		b.alloc = len(b.Buf) * Size
	}
}
