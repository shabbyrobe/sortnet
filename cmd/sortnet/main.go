package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
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
		return printPNG(net, outFile)
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

	var (
		xSize     = 20
		ySize     = 15
		dotRadius = 4
		width     = len(net.Ops)*xSize + xSize
		height    = net.Size*ySize + ySize

		bgCol   = color.RGBA{16, 16, 16, 255}
		slotCol = color.RGBA{96, 96, 96, 255}
		joinCol = color.RGBA{224, 224, 224, 255}
		dotCol  = color.RGBA{32, 160, 32, 255}

		palette = color.Palette{bgCol, slotCol, joinCol, dotCol}
	)

	im := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	draw.Draw(im, im.Bounds(), &image.Uniform{bgCol}, image.Point{}, draw.Src)

	// grid lines
	for i := 0; i < net.Size; i++ {
		y := (i*ySize + ySize)
		dims := image.Rect(xSize, y, width-xSize, y+1)
		draw.Draw(im, dims, &image.Uniform{slotCol}, image.Point{}, draw.Src)
	}

	for idx, c := range net.Ops {
		x := (idx*xSize + xSize)
		y1 := c.From*ySize + ySize
		y2 := c.To*ySize + ySize

		dims := image.Rect(x, y1, x+1, y2)
		draw.Draw(im, dims, &image.Uniform{joinCol}, image.Point{}, draw.Src)

		dot := &circle{image.Point{x, y1}, dotRadius, dotCol}
		draw.Draw(im, im.Bounds(), dot, image.Point{}, draw.Src)

		dot.p.Y = y2
		draw.Draw(im, im.Bounds(), dot, image.Point{}, draw.Src)
	}

	var w io.Writer

	if fileName == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
		w = f
	}

	if err := png.Encode(w, im); err != nil {
		return err
	}

	return nil
}

type circle struct {
	p image.Point
	r int
	c color.RGBA
}

func (c *circle) ColorModel() color.Model { return color.RGBAModel }

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return c.c
	}
	return color.RGBA{}
}
