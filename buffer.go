package block

import (
	"bytes"
	"fmt"
	"io"
)

// Buffer is a fixed-size block buffer that implements the io.Reader,
// io.Writer, io.Seeker, io.Closer, io.WriterAt, and io.ReaderAt interfaces.
type Buffer struct {
	Block
	p  int
	bp int
	bn int
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
	return b.Block.Len()
}

// Close closes the buffer. Further reads or writes panic.
func (b *Buffer) Close() error {
	return b.Block.Close()
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
		offset = int64(b.Len()) + offset
	}
	if offset < 0 {
		return int64(b.p), io.ErrUnexpectedEOF
	}
	b.p = int(offset)
	b.bn, b.bp = b.Block.Address(offset)
	return int64(b.p), nil
}

// Insert inserts p at offset q0. Shifting the
// rest on the right toward the infinity on the right.
// Insert is guaranteed to grow the buffer by len(p)
// bytes
func (b *Buffer) Insert(p []byte, q0 int64) (n int) {
	var err error
	growth := int64(len(p))
	b.Ensure(b.Len() + growth)
	defer func(n int64) {
		b.SetLen(n)
	}(b.Len() + growth)
	tmp := make([]byte, Size)
	defer func(q0 int64) {
		n, _ = b.WriteAt(p, q0)
	}(q0)
	q0 += int64(len(p))
	for q0-int64(len(p)) < b.Len() {
		defer func(q0 int64) {
			if err != nil {
				return
			}
			n, err = b.ReadAt(tmp, int64(q0))
			n, err = b.WriteAt(tmp[:n], q0+int64(len(p)))
		}(q0 - int64(len(p)))
		q0 += int64(len(tmp))
	}
	//	b.Len() += growth // mildly worked when here
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
	l := b.Len() - dx
	for at := int64(0); at+q0 != int64(b.Len()); {
		n, e1 = b.ReadAt(tmp, q1+at)
		n, e2 = b.WriteAt(tmp[:n], q0+at)
		at += int64(n)
		if e1 != nil || e2 != nil {
			break
		}
	}
	b.SetLen(l)
	return
}

// Write writes len(p) bytes into the buffer at at. It is an error
// to write to an offset beyond the buffer's current length.
func (b *Buffer) WriteAt(p []byte, at int64) (n int, err error) {
	bn, bp := b.Address(at)
	b.Ensure(at + int64(len(p)))
	for n < len(p) {
		m := b.WriteAddress(p[n:], bn, bp)
		n += m
		bp += m
		at += int64(m)
		if bp > Size {
			panic("bp > Size")
		}
		if bp == Size {
			bp = 0
			bn++
		}
	}
	if at > b.Len() {
		b.SetLen(at)
	}
	//	if bn > b.nblock {
	//		b.nblock = bn
	//	}
	return
}

// Read reads up to len(p) bytes into the slice until EOF startingat at.
func (b *Buffer) ReadAt(p []byte, at int64) (n int, err error) {
	bn, bp := b.Address(at)
	for {
		if bp == Size {
			bp = 0
			bn++
		}
		if at == b.Len() {
			return n, io.EOF
		}
		m := b.ReadAddress(p[n:], bn, bp)
		if m == 0 {
			return n, nil
		}
		n += m
		bp += m
		at += int64(m)
	}
	return
}

// Write writes len(p) bytes into the buffer. The buffer grows as necessary.
func (b *Buffer) Write(p []byte) (n int, err error) {
	n, err = b.WriteAt(p, int64(b.p))
	b.p += n
	return
}

// Read reads up to len(p) bytes into the slice until EOF
func (b *Buffer) Read(p []byte) (n int, err error) {
	n, err = b.ReadAt(p, int64(b.p))
	b.p += n
	return
}

// Dump is a debugging function that prints the output of the buffer
// in diagnostic form
func (b *Buffer) Dump() (s string) {
	s += b.Block.Dump()
	s += fmt.Sprintf("p: %d\n", b.p)
	s += fmt.Sprintf("bp: %d\n", b.bp)
	s += fmt.Sprintf("bn: %d\n", b.bn)
	return s
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
