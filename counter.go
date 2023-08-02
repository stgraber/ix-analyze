package main

type counter struct {
	hwaddr string
	v4rx   int64
	v4tx   int64
	v6rx   int64
	v6tx   int64
	total  int64
}

type trafficCounter map[string]*counter

func (tc trafficCounter) toSlice() []*counter {
	out := make([]*counter, 0, len(tc))
	for _, v := range tc {
		out = append(out, v)
	}

	return out
}
