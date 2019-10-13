package sortnet

type CompareAndSwap struct {
	From int
	To   int
}

type Network struct {
	Kind string
	Ops  []CompareAndSwap
	Size int

	// Depth is defined (informally) as the largest number of comparators that any input
	// value can encounter on its way through the network.
	// Depth may not be calculated for certain networks, so it may be 0.
	Depth int
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
