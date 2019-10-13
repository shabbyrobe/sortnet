package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shabbyrobe/sortnet"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var alg string
	var outFmt string
	var outFile string
	var n int
	var showInfo bool

	flag.IntVar(&n, "n", 0, "Network size")
	flag.StringVar(&alg, "alg", "best", "Algorithm (best, bosenelson)")
	flag.BoolVar(&showInfo, "info", true, "Show extra info about network on stderr")
	flag.StringVar(&outFmt, "fmt", "swaps", "Output format (swaps, png)")
	flag.StringVar(&outFile, "o", "", "Output file (for png)")
	flag.Parse()

	var net sortnet.Network
	switch alg {
	case "bosenelson":
		net = sortnet.BoseNelson(int(n))
	case "best":
		net = sortnet.New(int(n))
	default:
		return fmt.Errorf("unknown algo")
	}

	if n < 1 {
		return fmt.Errorf("network size (-n) must be >= 1")
	}

	if showInfo {
		fmt.Fprintln(os.Stderr, "kind:", net.Kind, "depth:", net.Depth, "size:", net.Size)
		fmt.Fprintln(os.Stderr)
	}

	switch outFmt {
	case "swaps":
		return printSwaps(net)
	case "png":
		return printPNG(net, f)
	default:
		return fmt.Errorf("unknown output format")
	}
}

func printSwaps(net sortnet.Network) error {
	for _, c := range net.Ops {
		fmt.Printf("swap(%d, %d)\n", c.From, c.To)
	}
	return nil
}

func printPNG(net sortnet.Network, fileName string) error {
	if fileName == "" {
		return fmt.Errorf("missing output filename (-o) for png")
	}

	for _, c := range net.Ops {
		fmt.Printf("swap(%d, %d)\n", c.From, c.To)
	}
}
