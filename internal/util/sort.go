package util

import "sort"

type SortFunc struct {
	Lenf  func() int
	Lessf func(i, j int) bool
	Swapf func(i, j int)
}

var _ sort.Interface = SortFunc{}

func (s SortFunc) Len() int {
	return s.Lenf()
}

func (s SortFunc) Less(i, j int) bool {
	return s.Lessf(i, j)
}

func (s SortFunc) Swap(i, j int) {
	s.Swapf(i, j)
}
