package sortnet

var (
	bySize           = [32][]Network{}
	shallowestBySize = [32]Network{}
	noNetwork        Network
)

func init() {
	for _, net := range Optimized {
		bySize[net.Size] = append(bySize[net.Size], net)

		shallowest := shallowestBySize[net.Size]
		if shallowest.Depth == 0 || shallowest.Depth > net.Depth {
			shallowestBySize[net.Size] = net
		}
	}
}

func New(size int) (net Network) {
	if size < 32 {
		optimized := shallowestBySize[size]
		if optimized.Size > 0 {
			return optimized
		}
	}

	return BoseNelson(size)
}
