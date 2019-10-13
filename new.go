package sortnet

var (
	bySize         = [32][]Network{}
	simplestBySize = [32]Network{}
	noNetwork      Network
)

func init() {
	for _, net := range Optimized {
		bySize[net.Size] = append(bySize[net.Size], net)

		simplest := simplestBySize[net.Size]
		if simplest.Size == 0 ||
			len(net.Ops) < len(simplest.Ops) ||
			(len(net.Ops) == len(simplest.Ops) && net.Depth < simplest.Depth) {

			simplestBySize[net.Size] = net
		}
	}
}

func New(size int) (net Network) {
	if size < 32 {
		optimized := simplestBySize[size]
		if optimized.Size > 0 {
			return optimized
		}
	}

	return BoseNelson(size)
}
