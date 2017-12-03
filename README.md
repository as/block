```
PACKAGE DOCUMENTATION

import "github.com/as/block"


const (
    // Size is the fixed size of each block in a buffer
    Size = 512
)

TYPES

type Buffer struct {
    Buf [][Size]byte
    // contains filtered or unexported fields
}
    Buffer is a fixed-size block buffer that implements the io.Reader,
    io.Writer, io.Seeker, io.Closer, io.WriterAt, and io.ReaderAt
    interfaces.

func (b *Buffer) Bytes() []byte
    Bytes returns a concatenated list of blocks as a slice. Very expensive
    and fancy. Don't use this.

func (b *Buffer) Close() error
    Close closes the buffer. Further reads or writes panic.

func (b *Buffer) Delete(q0, q1 int64) (n int)
    Delete deletes the bytes between [q0, q1). Shifting the rest on the
    right toward the zero offset on the left. Delete is a no-op if q0==q1.

func (b *Buffer) Dump() (s string)
    Dump is a debugging function that prints the output of the buffer in
    diagnostic form

func (b *Buffer) Insert(p []byte, q0 int64) (n int)
    Insert inserts p at offset q0. Shifting the rest on the right toward the
    infinity on the right. Insert is guaranteed to grow the buffer by len(p)
    bytes

func (b *Buffer) Len() int64
    Len returns the current length of the buffer

func (b *Buffer) Read(p []byte) (n int, err error)
    Read reads up to len(p) bytes into the slice until EOF

func (b *Buffer) ReadAt(p []byte, at int64) (n int, err error)
    Read reads up to len(p) bytes into the slice until EOF starting at at.

func (b *Buffer) Seek(offset int64, whence int) (int64, error)
    Seek seeks to the given offset using the rules of io.Seeker. It is not
    an error to seek to an offset greater than the length of the buffer,
    however, reads and writes after such a seek are undefined.

func (b *Buffer) String() string
    Strings returns the buffer as a string. Even fancier that Bytes().

func (b *Buffer) Write(p []byte) (n int, err error)
    Write writes len(p) bytes into the buffer. The buffer grows as
    necessary.

func (b *Buffer) WriteAt(p []byte, at int64) (n int, err error)
    Write writes len(p) bytes into the buffer at at. It is an error to write
    to an offset beyond the buffer's current length.

```
