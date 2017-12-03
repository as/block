package block

import (
	"testing"
)

func wr(b *Buffer, s string) (n int, err error) {
	return b.Write([]byte(s))
}
func in(b *Buffer, s string, at int) (n int) {
	return b.Insert([]byte(s), int64(at))
}
func del(b *Buffer, q0, q1 int) (n int) {
	return b.Delete(int64(q0), int64(q1))
}
func cklen(t *testing.T, b *Buffer, want int) {
	if b.Len() != int64(want) {
		t.Log(b.Dump())
		t.Logf("bad length: want %d have %d", want, b.Len())
		t.FailNow()
	}
}
func ckdata(t *testing.T, b *Buffer, want string) {
	if have := b.String(); have != want {
		t.Log(b.Dump())
		t.Logf("bad data: want %q have %q", want, have)
		t.FailNow()
	}
}

func TestInsert(t *testing.T) {
	b := new(Buffer)
	in(b, "the", 0)
	cklen(t, b, 3)
	in(b, " quick", 3)
	cklen(t, b, 3+6)
	in(b, " brown", 3+6)
	cklen(t, b, 3+6+6)
	in(b, " fox", 3+6+6)
	cklen(t, b, 3+6+6+4)
	ckdata(t, b, "the quick brown fox")
	in(b, "the", 0)
	in(b, " quick", 3)
	in(b, " brown", 3+6)
	in(b, " fox", 3+6+6)
	ckdata(t, b, "the quick brown foxthe quick brown fox")
	in(b, " fox", 0)
	ckdata(t, b, " foxthe quick brown foxthe quick brown fox")
}

func TestDelete(t *testing.T) {
	b := new(Buffer)
	in(b, "the quick brown fox jumps over the lazy dog", 0)
	ckdata(t, b, "the quick brown fox jumps over the lazy dog")
	del(b, 1, 2)
	ckdata(t, b, "te quick brown fox jumps over the lazy dog")
	del(b, 0, 0)
	ckdata(t, b, "te quick brown fox jumps over the lazy dog")
	del(b, 0, 1)
	ckdata(t, b, "e quick brown fox jumps over the lazy dog")
	del(b, 0, 17)
	ckdata(t, b, " jumps over the lazy dog")
	del(b, 0, 17)
	ckdata(t, b, "azy dog")
}

func TestWrite(t *testing.T) {
	b := new(Buffer)
	for _, v := range []string{"the quick ", "brown", " f", "", "o", "", "", "", "x", " jumps over ", "the lazy", " dog"} {
		n, err := wr(b, v)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if n != len(v) {
			t.Logf("bad length: %d\n", n)
			t.Fail()
		}
	}
	ckdata(t, b, "the quick brown fox jumps over the lazy dog")

}

func TestWriteAt(t *testing.T) {
	b := new(Buffer)
	x := "xoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxo"
	wr(b, x)
	b.Seek(0, 0)
	for i := 0; i < len(x); i++ {
		b.Write([]byte{'!'})
		b.WriteAt([]byte{'?'}, int64(i))
	}
	ckdata(t, b, "????????????????????????????????????????????????????????????????????????????????")
}

func TestSeek(t *testing.T) {
	b := new(Buffer)
	x := "xoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxoxo"
	wr(b, x)
	b.Seek(0, 0)
	for i := 1; i < len(x)-1; i++ {
		b.Seek(-int64(i), 2)
		b.Write([]byte{byte(i + 40)})
	}
	ckdata(t, b, "xovutsrqponmlkjihgfedcba`_^]\\[ZYXWVUTSRQPONMLKJIHGFEDCBA@?>=<;:9876543210/.-,+*)")
}
