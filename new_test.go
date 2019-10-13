package sortnet

import "testing"

func TestNew(t *testing.T) {
	for i := 0; i < 64; i++ {
		net := New(i)
		if net.Size != i {
			t.Fatal()
		}
	}
}
