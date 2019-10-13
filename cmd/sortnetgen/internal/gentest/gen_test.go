package gentest

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func BenchmarkSortNetInts(b *testing.B) {
	rng := rand.New(rand.NewSource(0))
	ints := newRandInts(rng, 10000000, 1024)

	for idx, tc := range []struct {
		name   string
		sz     int
		sorter func([]int)
	}{
		{"", 2, NetworkSort2xInt},
		{"", 3, NetworkSort3xInt},
		{"", 4, NetworkSort4xInt},
		{"", 5, NetworkSort5xInt},
		{"", 6, NetworkSort6xInt},
		{"", 7, NetworkSort7xInt},
		{"", 8, NetworkSort8xInt},
		{"", 9, NetworkSort9xInt},
		{"", 10, NetworkSort10xInt},
		{"", 11, NetworkSort11xInt},
		{"", 12, NetworkSort12xInt},
		{"", 13, NetworkSort13xInt},
		{"", 14, NetworkSort14xInt},
		{"", 15, NetworkSort15xInt},
		{"", 16, NetworkSort16xInt},
		{"", 24, NetworkSort24xInt},
		{"", 32, NetworkSort32xInt},
		{"", 48, NetworkSort48xInt},
		{"", 64, NetworkSort64xInt},
	} {
		_ = idx
		b.Run(fmt.Sprintf("network-%d", tc.sz), func(b *testing.B) {
			ints.Reset(b)
			for i := 0; i < b.N; i++ {
				cur := ints.Take(b, tc.sz)
				tc.sorter(cur)
			}
		})

		b.Run(fmt.Sprintf("std-%d", tc.sz), func(b *testing.B) {
			ints.Reset(b)
			for i := 0; i < b.N; i++ {
				cur := ints.Take(b, tc.sz)
				sort.Ints(cur)
			}
		})

		b.Run(fmt.Sprintf("stdslice-%d", tc.sz), func(b *testing.B) {
			ints.Reset(b)
			for i := 0; i < b.N; i++ {
				cur := ints.Take(b, tc.sz)
				sort.Slice(cur, func(i, j int) bool {
					return cur[i] < cur[j]
				})
			}
		})
	}
}

type randInts struct {
	rand *rand.Rand
	vs   []int
	next int
	sz   int
}

func newRandInts(r *rand.Rand, sz int, max int) *randInts {
	ints := &randInts{
		rand: r,
		vs:   make([]int, sz),
		sz:   sz,
	}
	for i := 0; i < sz; i++ {
		ints.vs[i] = r.Intn(max)
	}
	return ints
}

func (r *randInts) Reset(b *testing.B) {
	b.StopTimer()
	r.rand.Shuffle(r.sz, func(i, j int) {
		r.vs[i], r.vs[j] = r.vs[j], r.vs[i]
	})
	r.next = 0
	b.StartTimer()
}

func (r *randInts) Take(b *testing.B, n int) []int {
	if r.next+n >= r.sz {
		r.Reset(b)
	}
	out := r.vs[r.next : r.next+n]
	r.next += n
	return out
}
