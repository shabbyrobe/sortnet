package sortnet

func BoseNelson(n int) Network {
	var builder = boseNelsonBuilder{
		Network: Network{Kind: "Bose-Nelson", Size: n},
	}
	builder.split(0, n)
	return builder.Network
}

type boseNelsonBuilder struct {
	Network
}

func (builder *boseNelsonBuilder) split(first int, n int) {
	if n > 1 {
		mid := n / 2
		builder.split(first, mid)
		builder.split(first+mid, n-mid)
		builder.merge(first, mid, first+mid, n-mid)
	}
}

func (builder *boseNelsonBuilder) merge(s1i, s1len, s2i, s2len int) {
	if s1len == 1 && s2len == 1 {
		builder.Ops = append(builder.Ops, CompareAndSwap{s1i, s2i})

	} else if s1len == 1 && s2len == 2 {
		builder.Ops = append(builder.Ops, CompareAndSwap{s1i, s2i + 1})
		builder.Ops = append(builder.Ops, CompareAndSwap{s1i, s2i})

	} else if s1len == 2 && s2len == 1 {
		builder.Ops = append(builder.Ops, CompareAndSwap{s1i, s2i})
		builder.Ops = append(builder.Ops, CompareAndSwap{s1i + 1, s2i})

	} else {
		s1mid := s1len / 2
		s2mid := 0
		if s2len&1 == 0 {
			s2mid = s2len / 2
		} else {
			s2mid = (s2len + 1) / 2
		}
		builder.merge(s1i, s1mid, s2i, s2mid)
		builder.merge(s1i+s1mid, s1len-s1mid, s2i+s2mid, s2len-s2mid)
		builder.merge(s1i+s1mid, s1len-s1mid, s2i, s2mid)
	}
}
