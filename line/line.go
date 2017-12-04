package line

import "github.com/as/block"

type Counter struct {
	block.Block
	lines []int
}

func (b *Counter) Lines(block int) int {
	return b.lines[block]
}
func (b *Counter) WriteAddress(p []byte, block, offset int) int {
	ch := make(chan int)
	go func() {
		nl := 0
		for _, c := range p {
			if c == '\n' {
				nl++
			}
		}
		ch <- nl
	}()
	n := b.Block.WriteAddress(p, block, offset)
	if need := block + 1 - len(b.lines); need > 0 {
		b.lines = append(b.lines, make([]int, need, need+25)...)
	}
	b.lines[block] = <-ch
	return n
}
