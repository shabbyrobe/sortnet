package sortnet

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func maxNetSize(nets []Network) int {
	var max int
	for _, net := range nets {
		if net.Size > max {
			max = net.Size
		}
	}
	return max
}

func TestNetworks(t *testing.T) {
	var networks []Network

	for i := 1; i < 32; i++ {
		networks = append(networks, BoseNelson(i))
	}
	networks = append(networks, BoseNelson(64), BoseNelson(128))
	networks = append(networks, Optimized...)
	max := maxNetSize(networks)

	repeats := 1000
	rands := make([]int, max)
	rand.Seed(time.Now().UnixNano())

	for _, net := range networks {
		t.Run(fmt.Sprintf("%s(%d)", net.Kind, net.Size), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatal(r)
				}
			}()

			for repeat := 0; repeat < repeats; repeat++ {
				for i := 0; i < net.Size; i++ {
					// Lower number gives us a fighting chance of getting some dupes sometimes:
					rands[i] = int(rand.Int31n(1024))
				}

				stdSorted := make([]int, net.Size)
				copy(stdSorted, rands)

				netSorted := make([]int, net.Size)
				copy(netSorted, rands)

				sort.Ints(stdSorted)
				net.SortInts(netSorted)
				if !reflect.DeepEqual(stdSorted, netSorted) {
					t.Fatal(sortDiffMsg(net, stdSorted, netSorted))
				}

				sort.Slice(stdSorted, func(i, j int) bool { return stdSorted[i] > stdSorted[j] })
				net.SortIntsReverse(netSorted)

				if !reflect.DeepEqual(stdSorted, netSorted) {
					t.Fatal(sortDiffMsg(net, stdSorted, netSorted))
				}
			}
		})
	}
}

func sortDiffMsg(net Network, exp, out []int) string {
	expm := []string{}
	outm := []string{}
	for p := 0; p < net.Size; p++ {
		expm = append(expm, fmt.Sprintf("%-5d", exp[p]))
		outm = append(outm, fmt.Sprintf("%-5d", out[p]))
	}
	msg := fmt.Sprintf("\nexp: %s\nout: %s\n", strings.Join(expm, " "), strings.Join(outm, " "))
	return msg
}

var (
	BenchNetwork    Network
	BenchSortedInts []int
)

func BenchmarkBoseNelsonBuild(b *testing.B) {
	sizes := []int{
		1, 3, 16, 32, 64, 128, 1024,
	}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("%d", sz), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				BenchNetwork = BoseNelson(sz)
			}
		})
	}
}

func BenchmarkOptimized(b *testing.B) {
	// This benchmark is useless; the overhead of all the stuff you have to
	// do to get a set of unsorted ints for every benchmark iteration completely
	// overwhelms any meaningful difference in the algorithms. I think it's safe
	// to assume that "fewer ops, lower depth" == "better results", but I haven't
	// got a good way to test to make sure that's the case yet.

	nets := []Network{Green16, Senso16, VanVoorhis16, BoseNelson(16)}

	rng := rand.New(rand.NewSource(0))
	ints := newRandInts(rng, 10000000, 1024)

	for _, net := range nets {
		b.Run(fmt.Sprintf("%s(%d)", net.Kind, net.Size), func(b *testing.B) {
			ints.Reset(b)
			for i := 0; i < b.N; i++ {
				cur := ints.Take(b, net.Size)
				net.SortInts(cur)
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
