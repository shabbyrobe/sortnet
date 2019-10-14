package sortnet

import "reflect"

type CompareAndSwap struct {
	From int
	To   int
}

func (c CompareAndSwap) Reverse() CompareAndSwap {
	c.From, c.To = c.To, c.From // Yikes, this is a bit meta
	return c
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

// SortInts is a convenience that sorts a list of ints in place.
//
// This will be slower than the `sortnetgen` sorting network, but is still
// a fair bit faster than the stdlib for all the input sizes that have been
// tested (<64).
func (n Network) SortInts(vs []int) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortIntsReverse is a convenience that reverse sorts the input list in place.
// See SortInts for caveats.
func (n Network) SortIntsReverse(vs []int) {
	for _, c := range n.Ops {
		if vs[c.From] < vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortInt64s is a convenience that sorts a list of int64s in place.
// See SortInts for caveats.
func (n Network) SortInt64s(vs []int64) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortInt64sReverse is a convenience that reverse-sorts a list of int64s in place.
// See SortInts for caveats.
func (n Network) SortInt64sReverse(vs []int64) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortUint64s is a convenience that sorts a list of uint64s in place.
// See SortInts for caveats.
func (n Network) SortUint64s(vs []uint64) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortUint64sReverse is a convenience that reverse-sorts a list of uint64s in place.
// See SortInts for caveats.
func (n Network) SortUint64sReverse(vs []uint64) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortFloat64s is a convenience that sorts a list of float64s in place.
// See SortInts for caveats.
func (n Network) SortFloat64s(vs []float64) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortFloat64sReverse is a convenience that reverse-sorts a list of float64s in place.
// See SortInts for caveats.
func (n Network) SortFloat64sReverse(vs []float64) {
	for _, c := range n.Ops {
		if vs[c.From] > vs[c.To] {
			vs[c.From], vs[c.To] = vs[c.To], vs[c.From]
		}
	}
}

// SortSlice is a convenience that sorts the input list in place.
//
// Unlike SortInts, this is currently _substantially_ slower than the stdlib. It
// is not recommended for use at all, until and unless the egregious performance
// can be dealt with.
func (n Network) SortSlice(vs interface{}, less func(i, j int) bool) {
	v := reflect.ValueOf(vs)
	if v.Kind() != reflect.Slice {
		panic("value is not a slice")
	}
	if v.Len() == 0 {
		return
	}

	for _, c := range n.Ops {
		if less(c.From, c.To) {
			v1, v2 := v.Index(c.From), v.Index(c.To)
			tmp := v1.Interface()
			v1.Set(v2)
			v2.Set(reflect.ValueOf(tmp))
		}
	}
}
