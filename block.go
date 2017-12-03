package block

import (
	"bytes"
	"fmt"
	"io"
)

const (
	// Size is the fixed size of each block in a buffer
	Size = 512

	delta = 2
)

// Buffer is a fixed-size block buffer that implements the io.Reader,
// io.Writer, io.Seeker, io.Closer, io.WriterAt, and io.ReaderAt interfaces.
type Buffer struct {
	size   int
	Buf    [][Size]byte
	p      int
	bp     int
	bn     int
	len    int
	alloc  int
	nblock int
}

// Bytes returns a concatenated list of blocks as a slice.
// Very expensive and fancy. Don't use this.
func (b *Buffer) Bytes() []byte {
	var buf *bytes.Buffer
	func() {
		if buf != nil {
			return
		}
		buf = new(bytes.Buffer)
		at, _ := b.Seek(0, io.SeekCurrent)
		b.Seek(0, io.SeekStart)
		io.CopyN(buf, b, b.Len())
		b.Seek(at, io.SeekStart)
	}()
	return buf.Bytes()
}

// Len returns the current length of the buffer
func (b *Buffer) Len() int64 {
	return int64(b.len)
}

// Close closes the buffer. Further reads or writes panic.
func (b *Buffer) Close() error {
	b.Buf = nil
	return nil
}

// Strings returns the buffer as a string. Even fancier that Bytes().
func (b *Buffer) String() string {
	return string(b.Bytes())
}

// Seek seeks to the given offset using the rules of io.Seeker. It
// is not an error to seek to an offset greater than the length
// of the buffer, however, reads and writes after such a seek are
// undefined.
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

// Insert inserts p at offset q0. Shifting the
// rest on the right toward the infinity on the right.
// Insert is guaranteed to grow the buffer by len(p)
// bytes
func (b *Buffer) Insert(p []byte, q0 int64) (n int) {
	var err error
	growth := len(p)
	b.ensure(b.len + growth)
	defer func(n int) {
		b.len = n
	}(b.len + growth)
	tmp := make([]byte, Size)
	defer func(q0 int64) {
		n, _ = b.WriteAt(p, q0)
	}(q0)
	q0 += int64(len(p))
	for int(q0)-len(p) < b.len {
		defer func(q0 int64) {
			if err != nil {
				return
			}
			n, err = b.ReadAt(tmp, int64(q0))
			n, err = b.writeAt(tmp[:n], q0+int64(len(p)))
		}(q0 - int64(len(p)))
		q0 += int64(len(tmp))
	}
	b.len += growth // mildly worked when here
	return
}

// Delete deletes the bytes between [q0, q1). Shifting the
// rest on the right toward the zero offset on the left.
// Delete is a no-op if q0==q1.
func (b *Buffer) Delete(q0, q1 int64) (n int) {
	if q0 == q1 {
		return 0
	}
	var e1, e2 error
	tmp := make([]byte, Size)
	dx := q1 - q0
	l := b.len - int(dx)
	for at := int64(0); at+q0 != int64(b.len); {
		n, e1 = b.ReadAt(tmp, q1+at)
		n, e2 = b.WriteAt(tmp[:n], q0+at)
		at += int64(n)
		if e1 != nil || e2 != nil {
			break
		}
	}
	b.len = l
	return
}

// Write writes len(p) bytes into the buffer at at. It is an error
// to write to an offset beyond the buffer's current length.
func (b *Buffer) WriteAt(p []byte, at int64) (n int, err error) {
	if at > int64(b.len) {
		return 0, io.ErrUnexpectedEOF
	}
	return b.writeAt(p, at)
}

// Read reads up to len(p) bytes into the slice until EOF starting
// at at.
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

// Write writes len(p) bytes into the buffer. The buffer grows
// as necessary.
func (b *Buffer) Write(p []byte) (n int, err error) {
	//	msg("write request %q", p)
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
	if b.bn > b.nblock {
		b.nblock = b.bn
	}
	//	msg("write %d bytes", n)
	return
}

// Read reads up to len(p) bytes into the slice until EOF
func (b *Buffer) Read(p []byte) (n int, err error) {
	for {
		if b.bp == Size {
			b.bp = 0
			b.bn++
		}
		if b.p == b.len {
			return n, io.EOF
		}
		//		fmt.Println(b.Dump())
		m := copy(p[n:], b.Buf[b.bn][b.bp:])
		if m == 0 {
			return n, nil
		}
		n += m
		b.bp += m
		b.p += m
		//p = p[n:]
	}
	//	msg("read %d bytes", n)
	return
}

// Dump is a debugging function that prints the output of the buffer
// in diagnostic form
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

func (b *Buffer) ensure(n int) {
	if n > b.alloc {
		b.Buf = append(b.Buf, make([][Size]byte, n/Size+delta)...)
		b.alloc = len(b.Buf) * Size
	}
}

func (b *Buffer) findRegion(at int64) (p, bn, bp int) {
	if false {
		defer func() {
			fmt.Printf("findregion p/bn/bp: %d/%d/%d\n", p, bn, bp)
		}()
		fmt.Println(b.Dump())
	}
	return int(at), int(at / Size), int(at % Size)
}

func (b *Buffer) writeAt(p []byte, at int64) (n int, err error) {
	x, y, z := b.p, b.bn, b.bp

	b.p, b.bn, b.bp = b.findRegion(at)
	n, err = b.Write(p)

	b.p, b.bn, b.bp = x, y, z
	return n, err
}

func main() {
	bl := &Buffer{}
	bl.WriteAt([]byte("g"), 0)
	bl.WriteAt([]byte("o"), 1)
	bl.Insert([]byte("@"), 0)
	bl.Delete(0, 1)
	fmt.Println(bl.String())
}
func msg(fm string, i ...interface{}) {
	fmt.Printf(fm, i...)
	fmt.Println()
}
