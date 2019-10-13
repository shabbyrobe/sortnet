// Package sortnet provides algorithms and code generators for fast, zero-allocation
// sorting networks.
//
// The networks are perhaps most useful in the form of the `sortnetgen` code generator,
// found in github.com/shabbyrobe/sortnet/cmd/sortnetgen.
//
// Quickstart, generate sorting networks for strings from 2 to 8 items in both forward and
// reverse order:
//
//	go get github.com/shabbyrobe/sortnet/cmd/sortnetgen
//	sortnetgen -o - -pkg main -size 2-8 -fwd -rev string
//
package sortnet
