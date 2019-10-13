package sortnet

type CompareAndSwap struct {
	From int
	To   int
}

type Network struct {
	Ops   []CompareAndSwap
	Size  int
	Depth int
	Kind  string
}

// SortInts is a convenience; sorting networks are really only useful
// for small slices of a fixed size with a generated sort function.
func (n Network) SortInts(vs []int) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

func (n Network) SortIntsWithSwaps(vs []int) (swaps int) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
			swaps++
		}
	}
	return swaps
}
