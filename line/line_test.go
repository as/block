package line

import (
	"testing"

	"github.com/as/block"
)

func TestSimpleCount(t *testing.T) {
	ctr := &Counter{Block: &block.Fixed{}}
	b := block.Buffer{Block: ctr}
	b.Write([]byte("This\nis\na\ntest"))
	want := 3
	have := ctr.Lines(0)
	if want != have {
		t.Logf("want %d, have %d\n", want, have)
		t.Fail()
	}
}
