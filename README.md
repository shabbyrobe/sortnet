Sorting Network Generator for Go
================================

Fast zero-allocation sorting networks for Go. Useful when attempting to sort
easily compared values in very short lists. The sorting networks remain faster
than stdlib sort up to at least 128 items, but the generated code starts to
get quite large after about 32. YMMV.

Those "64 bytes-per-op" in the benchmark allocations might not seem like much,
but this was put together to help out with some color sorting being done on
every single pixel in an image; 64 bytes * 3 colour channels * 5 megapixels per
image starts to get a bit more antagonistic towards the garbage collector!

![sortnet30](https://raw.githubusercontent.com/shabbyrobe/sortnet/master/assets/sortnet30-sml.png)

_Each horizontal line represents an array index, and each vertical line represents
a "compare and swap" operation between those two indices_

There are two tools in this repo:

- `github.com/shabbyrobe/sortnet/cmd/sortnet`: Miscellaneous tools for playing
  with sorting networks. Used to generate the above image.

- `github.com/shabbyrobe/sortnet/cmd/sortnetgen`: Command line tool for use
  with `go generate`; generates code for sorting networks of arbitrary types.

Install:

    go get github.com/shabbyrobe/sortnet/cmd/sortnetgen

If running using `go:generate`, it's not necessary to pass `-pkg`, but if running from
the command line, it's required. Examples below will presume `-pkg` is _not_ required.

Generate forward and reverse sorting network of sizes 3-5 for float64:

    sortnetgen -fwd -rev -size 3-5 float64

Generate forward sorting network of sizes 3-5 for int64

    sortnetgen -fwd -size 3-5 int64

Generate reverse sorting network of sizes 3, 4, 5 and 9 for string:

    sortnetgen -rev size 3-5, string

The type will be the basis for the comparison. If `<input>` is a builtin primitive, `<` is
used for comparisons, otherwise -greater and -less are used to determine how to compare
and swap for -fwd and -rev sorts respectively.

Generate forward sorting network of size 2 for `example.com/foo.Yep`, providing `-greater`:

	sortnetgen -size 2 -greater 'foo.YepCASGreater(&a[{{.From}}], &a[{{.To}}])' example.com/foo.Yep

If neither `{{.From}}` nor `{{.To}}` are present in the template, it is presumed to
be a function. The following is equivalent to the previous example:

	sortnetgen -size 2 -greater 'foo.YepCASGreater' example.com/foo.Yep


Crappy Benchmarks Game
----------------------

For a slice of ints, a sorting network handily beats the standard library by
about 4x. Benchmarks are on a fairly ancient dual-core 2015 MacBook Pro.
The `network` benchmarks show a generated sorting network, `std` shows the
stdlib's `sort.Ints` function, and `stdslice` shows `sort.Slice`:

    BenchmarkSortNetInts/network-3-4         	 5415687	        23.0 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-3-4             	 1200658	        96.2 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-3-4        	  726608	       167 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-4-4         	 3500810	        34.9 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-4-4             	  987211	       124 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-4-4        	  698678	       195 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-5-4         	 2821161	        48.1 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-5-4             	  797575	       161 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-5-4        	  567574	       217 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-6-4         	 2147130	        51.5 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-6-4             	  642166	       191 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-6-4        	  508189	       241 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-7-4         	 1880780	        64.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-7-4             	  523125	       243 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-7-4        	  439960	       271 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-8-4         	 1554298	        77.8 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-8-4             	  443148	       297 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-8-4        	  382012	       337 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-16-4        	  594891	       215 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-16-4            	  162730	       791 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-16-4       	  160850	       792 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-24-4        	  327681	       402 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-24-4            	   96368	      1406 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-24-4       	   98757	      1292 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-32-4        	  183176	       668 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-32-4            	   58293	      1947 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-32-4       	   65408	      1744 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-48-4        	   73920	      1433 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-48-4            	   37552	      3419 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-48-4       	   41648	      3028 ns/op	      64 B/op	       2 allocs/op
    BenchmarkSortNetInts/network-64-4        	   60121	      2029 ns/op	       0 B/op	       0 allocs/op
    BenchmarkSortNetInts/std-64-4            	   26676	      4892 ns/op	      32 B/op	       1 allocs/op
    BenchmarkSortNetInts/stdslice-64-4       	   29008	      4227 ns/op	      64 B/op	       2 allocs/op

